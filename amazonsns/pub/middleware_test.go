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
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

var (
	dummyTraceID = trace.TraceID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x21, 0xdc, 0x18, 0x7, 0x52, 0x47, 0x85}
	dummySpanID  = trace.SpanID{0x0, 0x21, 0x97, 0xec, 0x5d, 0x8a, 0x25, 0xe}
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
	wantMsgAttrs := map[string]messageAttributeValue{
		"traceparent": {DataType: utils.DataTypeString, Value: fmt.Sprintf("00-%s-%s-01", span.SpanContext().TraceID(), span.SpanContext().SpanID())},
	}
	if diff := cmp.Diff(wantMsgAttrs, gotMsgAttrs); diff != "" {
		t.Errorf("message attributes (-want, +got):\n%s", diff)
	}
	wantSpans := []tracetest.SpanStub{
		{
			Name:        "parent",
			SpanContext: trace.NewSpanContext(trace.SpanContextConfig{TraceFlags: trace.FlagsSampled, TraceID: dummyTraceID, SpanID: dummySpanID}),
			SpanKind:    trace.SpanKindInternal,
			Resource: resource.NewSchemaless(
				attribute.String("service.name", "unknown_service:pub.test"),
				attribute.String("telemetry.sdk.language", "go"),
				attribute.String("telemetry.sdk.name", "opentelemetry"),
				attribute.String("telemetry.sdk.version", "1.43.0"),
			),
			InstrumentationScope: instrumentation.Scope{
				Name: "test",
			},
		},
	}
	if diff := diffSpans(wantSpans, exporter.GetSpans()); diff != "" {
		t.Errorf("spans (-want, +got):\n%s", diff)
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
	wantSpans := []tracetest.SpanStub{
		{
			Name:        "parent",
			SpanContext: trace.NewSpanContext(trace.SpanContextConfig{TraceFlags: trace.FlagsSampled, TraceID: dummyTraceID, SpanID: dummySpanID}),
			SpanKind:    trace.SpanKindInternal,
			Resource: resource.NewSchemaless(
				attribute.String("service.name", "unknown_service:pub.test"),
				attribute.String("telemetry.sdk.language", "go"),
				attribute.String("telemetry.sdk.name", "opentelemetry"),
				attribute.String("telemetry.sdk.version", "1.43.0"),
			),
			InstrumentationScope: instrumentation.Scope{
				Name: "test",
			},
		},
	}
	if diff := diffSpans(wantSpans, exporter.GetSpans()); diff != "" {
		t.Errorf("spans (-want, +got):\n%s", diff)
	}
}

func diffSpans(want, got []tracetest.SpanStub) string {
	return cmp.Diff(want, got,
		cmpopts.IgnoreFields(
			tracetest.SpanStub{},
			"StartTime", "EndTime",
			"InstrumentationLibrary",
		),
		cmp.Comparer(func(a, b attribute.Set) bool {
			return a.Equals(&b)
		}),
		cmp.Comparer(func(a, b *resource.Resource) bool {
			return a.Equal(b)
		}),
		cmp.Comparer(func(a, b trace.SpanContext) bool {
			return a.HasTraceID() == b.HasTraceID() &&
				a.HasSpanID() == b.HasSpanID() &&
				a.IsSampled() == b.IsSampled() &&
				a.IsRemote() == b.IsRemote() &&
				a.IsValid() == b.IsValid()
		}),
	)
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
