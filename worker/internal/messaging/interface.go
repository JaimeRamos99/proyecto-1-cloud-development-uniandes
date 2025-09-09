package messaging

import (
	"context"
)

// MessageQueue defines the interface for message queue operations
type MessageQueue interface {
	// ReceiveMessages receives messages from the queue with long polling
	ReceiveMessages(ctx context.Context, maxMessages int32, waitTimeSeconds int32) ([]*ReceivedMessage, error)

	// DeleteMessage removes a processed message from the queue
	DeleteMessage(ctx context.Context, receiptHandle string) error

	// Close closes the connection to the message queue
	Close() error
}

// VideoProcessingMessage represents a video processing message
type VideoProcessingMessage struct {
	S3Key string `json:"s3_key"`
}

// ReceivedMessage represents a message received from the queue
type ReceivedMessage struct {
	Body          string
	ReceiptHandle string
	MessageID     string
}
