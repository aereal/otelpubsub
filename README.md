[![status][ci-status-badge]][ci-status]
[![PkgGoDev][pkg-go-dev-badge]][pkg-go-dev]

# otelpubsub

OpenTelemetry trace context propagation for AWS SNS and SQS messages.

## Overview

This library propagates trace context through AWS SNS/SQS message attributes, enabling distributed tracing across message-based architectures.

When a message is published, the current span's trace ID and span ID are injected into message attributes. 

When the message is received, the trace context is extracted and linked to the processing span.

## Installation

```bash
go get github.com/aereal/otelpubsub/amazonsns # for Amazon SNS producers/consumers
go get github.com/aereal/otelpubsub/amazonsqs # for Amazon SQS producers/consumers
```

## Usage

### Publishing messages (SNS)

```go
import (
    "github.com/aws/aws-sdk-go-v2/service/sns"
    "github.com/aereal/otelpubsub/amazonsns/pub"
)

// Register middleware when creating the SNS client
client := sns.NewFromConfig(cfg, func(o *sns.Options) {
    pub.AppendMiddlewares(&o.APIOptions)
})

// Trace context is automatically injected into message attributes
client.Publish(ctx, &sns.PublishInput{
    TopicArn: &topicArn,
    Message:  &message,
})
```

### Publishing messages (SQS)

```go
import (
    "github.com/aws/aws-sdk-go-v2/service/sqs"
    "github.com/aereal/otelpubsub/amazonsqs/pub"
)

client := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
    pub.AppendMiddlewares(&o.APIOptions)
})

client.SendMessage(ctx, &sqs.SendMessageInput{
    QueueUrl:    &queueUrl,
    MessageBody: &message,
})
```

### Processing messages (SQS via Lambda)

```go
import (
    "github.com/aereal/otelpubsub/amazonsqs/sub"
)

// Wrap your processor to automatically create spans with trace links
processor := sqssub.WrapProcessor(func(ctx context.Context, msg *sub.Message) error {
    // Process message with trace context
    return nil
})
```

## License

See LICENSE file.

[pkg-go-dev]: https://pkg.go.dev/github.com/aereal/otelpubsub
[pkg-go-dev-badge]: https://pkg.go.dev/badge/github.com/aereal/otelpubsub.svg
[ci-status-badge]: https://github.com/aereal/otelpubsub/workflows/CI/badge.svg?branch=main
[ci-status]: https://github.com/aereal/otelpubsub/actions/workflows/CI
