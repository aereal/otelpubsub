package pub

import (
	"context"

	"github.com/aereal/otelpubsub"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/aws/smithy-go/middleware"
)

// AppendMiddlewares registers a middleware that injects trace context into SNS message attributes
// before Publish and PublishBatch API calls.
// Pass the APIOptions field from [sns.Options] to this function.
func AppendMiddlewares(apiOptions *[]func(*middleware.Stack) error) {
	*apiOptions = append(*apiOptions, func(stack *middleware.Stack) error {
		return stack.Initialize.Add(middleware.InitializeMiddlewareFunc("InstrumentPub", instrumentPublish), middleware.Before)
	})
}

func instrumentPublish(ctx context.Context, input middleware.InitializeInput, next middleware.InitializeHandler) (middleware.InitializeOutput, middleware.Metadata, error) {
	switch params := input.Parameters.(type) {
	case *sns.PublishInput:
		mas := params.MessageAttributes
		if mas == nil {
			mas = map[string]types.MessageAttributeValue{}
		}
		otelpubsub.Propagator{}.Inject(ctx, NewMessageAttributeCarrier(mas))
		params.MessageAttributes = mas
		input.Parameters = params
	case *sns.PublishBatchInput:
		entries := make([]types.PublishBatchRequestEntry, 0, len(params.PublishBatchRequestEntries))
		for _, original := range params.PublishBatchRequestEntries {
			entry := cloneEntry(original)
			if entry.MessageAttributes == nil {
				entry.MessageAttributes = map[string]types.MessageAttributeValue{}
			}
			otelpubsub.Propagator{}.Inject(ctx, NewMessageAttributeCarrier(entry.MessageAttributes))
			entries = append(entries, entry)
		}
		params.PublishBatchRequestEntries = entries
		input.Parameters = params
	}
	return next.HandleInitialize(ctx, input)
}

func cloneEntry(original types.PublishBatchRequestEntry) types.PublishBatchRequestEntry {
	return types.PublishBatchRequestEntry{
		Id:                     original.Id,
		Message:                original.Message,
		MessageAttributes:      original.MessageAttributes,
		MessageDeduplicationId: original.MessageDeduplicationId,
		MessageGroupId:         original.MessageGroupId,
		MessageStructure:       original.MessageStructure,
		Subject:                original.Subject,
	}
}
