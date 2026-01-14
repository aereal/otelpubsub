package sub

import "encoding/json"

type Message struct {
	Attributes             map[string]string `json:"attributes"`
	MessageAttributes      MessageAttributes `json:"messageAttributes"`
	MessageID              string            `json:"messageId"`
	ReceiptHandle          string            `json:"receiptHandle"`
	MD5OfBody              string            `json:"md5OfBody"`
	MD5OfMessageAttributes string            `json:"md5OfMessageAttributes"`
	EventSourceARN         string            `json:"eventSourceARN"`
	EventSource            string            `json:"eventSource"`
	AWSRegion              string            `json:"awsRegion"`
	Body                   json.RawMessage   `json:"body"`
}
