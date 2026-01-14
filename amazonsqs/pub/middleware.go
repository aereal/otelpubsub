package pub

import (
	"context"

	"github.com/aereal/otelpubsub/amazonsqs/internal"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/smithy-go/middleware"
)

// AppendMiddlewares registers a middleware that injects trace context into SQS message attributes
// before SendMessage and SendMessageBatch API calls.
// Pass the APIOptions field from [sqs.Options] to this function.
func AppendMiddlewares(apiOptions *[]func(*middleware.Stack) error) {
	*apiOptions = append(*apiOptions, func(stack *middleware.Stack) error {
		return stack.Initialize.Add(middleware.InitializeMiddlewareFunc("InstrumentPub", instrumentPublish), middleware.Before)
	})
}

func instrumentPublish(ctx context.Context, input middleware.InitializeInput, next middleware.InitializeHandler) (middleware.InitializeOutput, middleware.Metadata, error) {
	switch params := input.Parameters.(type) {
	case *sqs.SendMessageInput:
		mas := params.MessageAttributes
		if mas == nil {
			mas = map[string]types.MessageAttributeValue{}
		}
		internal.Propagator.Inject(ctx, NewMessageAttributeCarrier(mas))
		params.MessageAttributes = mas
		input.Parameters = params
	case *sqs.SendMessageBatchInput:
		entries := make([]types.SendMessageBatchRequestEntry, 0, len(params.Entries))
		for _, original := range params.Entries {
			entry := cloneEntry(original)
			if entry.MessageAttributes == nil {
				entry.MessageAttributes = map[string]types.MessageAttributeValue{}
			}
			internal.Propagator.Inject(ctx, NewMessageAttributeCarrier(entry.MessageAttributes))
			entries = append(entries, entry)
		}
		params.Entries = entries
		input.Parameters = params
	}
	return next.HandleInitialize(ctx, input)
}

func cloneEntry(original types.SendMessageBatchRequestEntry) types.SendMessageBatchRequestEntry {
	return types.SendMessageBatchRequestEntry{
		Id:                      original.Id,
		MessageBody:             original.MessageBody,
		DelaySeconds:            original.DelaySeconds,
		MessageAttributes:       original.MessageAttributes,
		MessageDeduplicationId:  original.MessageDeduplicationId,
		MessageGroupId:          original.MessageGroupId,
		MessageSystemAttributes: original.MessageSystemAttributes,
	}
}
