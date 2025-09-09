package messaging

import (
	"context"
)

// MessageQueue defines the interface for message queue operations
type MessageQueue interface {
	// SendMessage sends a message to the queue
	SendMessage(ctx context.Context, message Message) error
	
	// Close closes the connection to the message queue
	Close() error
}

// Message represents a generic message structure
type Message interface {
	// GetBody returns the message body as string
	GetBody() string
}

// VideoProcessingMessage represents a video processing message
type VideoProcessingMessage struct {
	S3Key string `json:"s3_key"`
}

// GetBody implements Message interface - returns JSON representation
func (v *VideoProcessingMessage) GetBody() string {
	return "" // Will be serialized by the queue implementation
}
