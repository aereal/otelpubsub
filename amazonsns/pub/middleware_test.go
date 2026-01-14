package pub_test

import (
	stdcmp "cmp"
	"context"
	"fmt"
	"iter"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/aereal/otelpubsub/amazonsns/internal/utils"
	"github.com/aereal/otelpubsub/amazonsns/pub"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/google/go-cmp/cmp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type messageAttributeValue struct {
	DataType string
	Value    string
}

func TestMiddleware_publish(t *testing.T) {
	t.Parallel()

	var gotMsgAttrs map[string]messageAttributeValue
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %s", err)
			return
		}
		gotMsgAttrs = aggregateMessageAttributeValues(iterateSortedMapEntries(r.PostForm))
	}))
	t.Cleanup(srv.Close)
	cfg := aws.Config{
		Region:       "us-east-1",
		Credentials:  staticCredentials("id", "secret", "token"),
		BaseEndpoint: &srv.URL,
	}
	pub.AppendMiddlewares(&cfg.APIOptions)
	client := sns.NewFromConfig(cfg)

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))

	input := &sns.PublishInput{
		Message: utils.Ptr("msg"),
	}

	ctx, span := tp.Tracer("test").Start(t.Context(), "parent")
	if _, err := client.Publish(ctx, input); err != nil {
		t.Fatal(err)
	}
	span.End()
	if err := tp.ForceFlush(t.Context()); err != nil {
		t.Fatal(err)
	}
	gotSpans := exporter.GetSpans()
	t.Logf("%d spans got", len(gotSpans))
	for i, span := range gotSpans {
		t.Logf("span: index=%d name=%s", i, span.Name)
	}
	t.Logf("%d message attributes", len(gotMsgAttrs))
	wantMsgAttrs := map[string]messageAttributeValue{
		"traceparent": {DataType: utils.DataTypeString, Value: fmt.Sprintf("00-%s-%s-01", span.SpanContext().TraceID(), span.SpanContext().SpanID())},
	}
	if diff := cmp.Diff(wantMsgAttrs, gotMsgAttrs); diff != "" {
		t.Errorf("message attributes (-want, +got):\n%s", diff)
	}
}

func TestMiddleware_publish_batch(t *testing.T) {
	t.Parallel()

	gotMsgAttrs := map[string]map[string]messageAttributeValue{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %s", err)
			return
		}
		for entryIdx, members := range iterateMembers(iterateSortedMapEntries(r.PostForm)) {
			got := aggregateMessageAttributeValues(iterateSortedMapEntries(members))
			gotMsgAttrs[entryIdx] = got
		}
	}))
	t.Cleanup(srv.Close)
	cfg := aws.Config{
		Region:       "us-east-1",
		Credentials:  staticCredentials("id", "secret", "token"),
		BaseEndpoint: &srv.URL,
	}
	pub.AppendMiddlewares(&cfg.APIOptions)
	client := sns.NewFromConfig(cfg)

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))

	input := &sns.PublishBatchInput{
		TopicArn: utils.Ptr("arn:aws:sns:us-east-1:1234567890123:topic-1"),
		PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
			{
				Id:      utils.Ptr("1"),
				Message: utils.Ptr("msg-1"),
			},
			{
				Id:      utils.Ptr("2"),
				Message: utils.Ptr("msg-2"),
			},
		},
	}
	ctx, span := tp.Tracer("test").Start(t.Context(), "parent")
	if _, err := client.PublishBatch(ctx, input); err != nil {
		t.Fatal(err)
	}
	span.End()
	if err := tp.ForceFlush(t.Context()); err != nil {
		t.Fatal(err)
	}
	gotSpans := exporter.GetSpans()
	t.Logf("%d spans got", len(gotSpans))
	for i, span := range gotSpans {
		t.Logf("span: index=%d name=%s", i, span.Name)
	}
	t.Logf("%d message attributes", len(gotMsgAttrs))
	wantMsgAttrs := map[string]map[string]messageAttributeValue{
		"1": {
			"traceparent": {DataType: utils.DataTypeString, Value: fmt.Sprintf("00-%s-%s-01", span.SpanContext().TraceID(), span.SpanContext().SpanID())},
		},
		"2": {
			"traceparent": {DataType: utils.DataTypeString, Value: fmt.Sprintf("00-%s-%s-01", span.SpanContext().TraceID(), span.SpanContext().SpanID())},
		},
	}
	if diff := cmp.Diff(wantMsgAttrs, gotMsgAttrs); diff != "" {
		t.Errorf("message attributes (-want, +got):\n%s", diff)
	}
}

func staticCredentials(keyID, secret, sessionToken string) *awsCredentials {
	return &awsCredentials{Credentials: aws.Credentials{
		AccessKeyID:     keyID,
		SecretAccessKey: secret,
		SessionToken:    sessionToken,
	}}
}

type awsCredentials struct {
	aws.Credentials
}

var _ aws.CredentialsProvider = (*awsCredentials)(nil)

func (c *awsCredentials) Retrieve(_ context.Context) (aws.Credentials, error) {
	return c.Credentials, nil
}

func iterateSortedMapEntries[K stdcmp.Ordered, V any](m map[K]V) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, k := range slices.Sorted(maps.Keys(m)) {
			if !yield(k, m[k]) {
				return
			}
		}
	}
}

func aggregateMessageAttributeValues(pairs iter.Seq2[string, []string]) map[string]messageAttributeValue {
	idx2name := map[string]string{}
	ret := map[string]messageAttributeValue{}
	for key, values := range pairs {
		rest, ok := strings.CutPrefix(key, "MessageAttributes.entry.")
		if !ok {
			continue
		}
		idx, field, ok := strings.Cut(rest, ".")
		if !ok {
			continue
		}
		switch field {
		case "Name":
			name := values[0]
			idx2name[idx] = name
			ret[name] = messageAttributeValue{}
		case "Value.DataType":
			name, ok := idx2name[idx]
			if !ok {
				continue
			}
			av := ret[name]
			av.DataType = values[0]
			ret[name] = av
		case "Value.StringValue":
			name, ok := idx2name[idx]
			if !ok {
				continue
			}
			av := ret[name]
			av.Value = values[0]
			ret[name] = av
		}
	}
	return ret
}

func iterateMembers(pairs iter.Seq2[string, []string]) iter.Seq2[string, map[string][]string] {
	return func(yield func(string, map[string][]string) bool) {
		ret := map[string]map[string][]string{}
		for key, values := range pairs {
			after, ok := strings.CutPrefix(key, "PublishBatchRequestEntries.member.")
			if !ok {
				continue
			}
			idx, field, ok := strings.Cut(after, ".")
			if !ok {
				continue
			}
			if _, ok := ret[idx]; !ok {
				ret[idx] = map[string][]string{}
			}
			ret[idx][field] = values
		}
		iterateSortedMapEntries(ret)(yield)
	}
}
