package providers

import (
	"context"
	"fmt"

	"worker/internal/config"
	"worker/internal/messaging"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// SQSQueue implements the MessageQueue interface using AWS SQS
type SQSQueue struct {
	client   *sqs.Client
	queueURL string
}

// NewSQSQueue creates a new SQS-based message queue
func NewSQSQueue(cfg *config.AWSConfig) (*SQSQueue, error) {
	// Configure AWS config options
	var configOptions []func(*awsconfig.LoadOptions) error

	// Add region
	configOptions = append(configOptions, awsconfig.WithRegion(cfg.Region))

	// Only add static credentials if they are provided, otherwise use default credential chain (instance profile)
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		configOptions = append(configOptions, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			),
		))
	}

	// Load AWS config
	awsConfig, err := awsconfig.LoadDefaultConfig(context.TODO(), configOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create SQS client
	var client *sqs.Client
	if cfg.EndpointURL != "" {
		// LocalStack configuration - use custom endpoint
		client = sqs.NewFromConfig(awsConfig, func(o *sqs.Options) {
			o.BaseEndpoint = aws.String(cfg.EndpointURL)
		})
	} else {
		// Real AWS configuration
		client = sqs.NewFromConfig(awsConfig)
	}

	// Get queue URL
	queueURLResult, err := client.GetQueueUrl(context.TODO(), &sqs.GetQueueUrlInput{
		QueueName: aws.String(cfg.SQSQueueName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get queue URL: %w", err)
	}

	return &SQSQueue{
		client:   client,
		queueURL: aws.ToString(queueURLResult.QueueUrl),
	}, nil
}

// ReceiveMessages receives messages from the SQS queue with long polling
func (s *SQSQueue) ReceiveMessages(ctx context.Context, maxMessages int32, waitTimeSeconds int32) ([]*messaging.ReceivedMessage, error) {
	// SQS maximum wait time is 20 seconds
	if waitTimeSeconds > 20 {
		waitTimeSeconds = 20
	}

	// SQS maximum messages per request is 10
	if maxMessages > 10 {
		maxMessages = 10
	}

	result, err := s.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.queueURL),
		MaxNumberOfMessages: maxMessages,
		WaitTimeSeconds:     waitTimeSeconds,
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages from SQS: %w", err)
	}

	// Convert SQS messages to our internal format
	messages := make([]*messaging.ReceivedMessage, len(result.Messages))
	for i, msg := range result.Messages {
		messages[i] = &messaging.ReceivedMessage{
			Body:          aws.ToString(msg.Body),
			ReceiptHandle: aws.ToString(msg.ReceiptHandle),
			MessageID:     aws.ToString(msg.MessageId),
		}
	}

	return messages, nil
}

// DeleteMessage removes a processed message from the SQS queue
func (s *SQSQueue) DeleteMessage(ctx context.Context, receiptHandle string) error {
	_, err := s.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	if err != nil {
		return fmt.Errorf("failed to delete message from SQS: %w", err)
	}

	return nil
}

// Close closes the connection to the message queue
func (s *SQSQueue) Close() error {
	// SQS doesn't require explicit connection closing
	return nil
}
