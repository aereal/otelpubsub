package sub

// Event represents an SQS event delivered to AWS Lambda via event source mapping.
type Event struct {
	Records []Message `json:"records"`
}
