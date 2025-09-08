package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	appconfig "proyecto1/root/internal/config"
)

// SQSService handles SQS operations
type SQSService struct {
	client   *sqs.Client
	queueURL string
}

// Message represents an SQS message
type Message struct {
	ID           string            `json:"id"`
	Body         string            `json:"body"`
	Attributes   map[string]string `json:"attributes,omitempty"`
	ReceiptHandle string           `json:"receipt_handle,omitempty"`
}

// VideoProcessingMessage represents a video processing message
type VideoProcessingMessage struct {
	VideoID   string `json:"video_id"`
	UserID    string `json:"user_id"`
	S3Key     string `json:"s3_key"`
	Action    string `json:"action"` // "process", "transcode", "thumbnail"
	Priority  int    `json:"priority"`
	Timestamp int64  `json:"timestamp"`
}

// NewSQSService creates a new SQS service instance
func NewSQSService(cfg *appconfig.AWSConfig) (*SQSService, error) {
	var awsConfig aws.Config
	var err error

	if cfg.EndpointURL != "" {
		// LocalStack configuration
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               cfg.EndpointURL,
				HostnameImmutable: true,
			}, nil
		})

		awsConfig, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(cfg.Region),
			config.WithEndpointResolverWithOptions(customResolver),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			)),
		)
	} else {
		// Real AWS configuration
		awsConfig, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			)),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sqs.NewFromConfig(awsConfig)

	// Get queue URL
	queueURLResult, err := client.GetQueueUrl(context.TODO(), &sqs.GetQueueUrlInput{
		QueueName: aws.String(cfg.SQSQueueName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get queue URL: %w", err)
	}

	return &SQSService{
		client:   client,
		queueURL: aws.ToString(queueURLResult.QueueUrl),
	}, nil
}

// SendMessage sends a message to SQS
func (s *SQSService) SendMessage(ctx context.Context, message interface{}) error {
	messageBody, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = s.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.queueURL),
		MessageBody: aws.String(string(messageBody)),
	})
	if err != nil {
		return fmt.Errorf("failed to send message to SQS: %w", err)
	}

	return nil
}

// SendVideoProcessingMessage sends a video processing message to SQS
func (s *SQSService) SendVideoProcessingMessage(ctx context.Context, msg *VideoProcessingMessage) error {
	return s.SendMessage(ctx, msg)
}

// ReceiveMessages receives messages from SQS
func (s *SQSService) ReceiveMessages(ctx context.Context, maxMessages int32, waitTimeSeconds int32) ([]*Message, error) {
	output, err := s.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.queueURL),
		MaxNumberOfMessages: maxMessages,
		WaitTimeSeconds:     waitTimeSeconds,
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameAll,
		},
		MessageAttributeNames: []string{"All"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages from SQS: %w", err)
	}

	var messages []*Message
	for _, sqsMsg := range output.Messages {
		msg := &Message{
			ID:            aws.ToString(sqsMsg.MessageId),
			Body:          aws.ToString(sqsMsg.Body),
			ReceiptHandle: aws.ToString(sqsMsg.ReceiptHandle),
			Attributes:    make(map[string]string),
		}

		// Convert attributes
		for key, value := range sqsMsg.Attributes {
			msg.Attributes[string(key)] = value
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// DeleteMessage deletes a message from SQS
func (s *SQSService) DeleteMessage(ctx context.Context, receiptHandle string) error {
	_, err := s.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	if err != nil {
		return fmt.Errorf("failed to delete message from SQS: %w", err)
	}

	return nil
}

// ChangeMessageVisibility changes the visibility timeout of a message
func (s *SQSService) ChangeMessageVisibility(ctx context.Context, receiptHandle string, visibilityTimeoutSeconds int32) error {
	_, err := s.client.ChangeMessageVisibility(ctx, &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(s.queueURL),
		ReceiptHandle:     aws.String(receiptHandle),
		VisibilityTimeout: visibilityTimeoutSeconds,
	})
	if err != nil {
		return fmt.Errorf("failed to change message visibility: %w", err)
	}

	return nil
}

// GetQueueAttributes gets queue attributes
func (s *SQSService) GetQueueAttributes(ctx context.Context) (map[string]string, error) {
	output, err := s.client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(s.queueURL),
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameAll,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get queue attributes: %w", err)
	}

	attributes := make(map[string]string)
	for key, value := range output.Attributes {
		attributes[string(key)] = value
	}

	return attributes, nil
}

// GetApproximateMessageCount gets the approximate number of messages in the queue
func (s *SQSService) GetApproximateMessageCount(ctx context.Context) (int, error) {
	attributes, err := s.GetQueueAttributes(ctx)
	if err != nil {
		return 0, err
	}

	countStr, exists := attributes["ApproximateNumberOfMessages"]
	if !exists {
		return 0, fmt.Errorf("ApproximateNumberOfMessages attribute not found")
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse message count: %w", err)
	}

	return count, nil
}
