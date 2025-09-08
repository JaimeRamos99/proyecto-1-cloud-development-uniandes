package providers

import (
	"context"
	"encoding/json"
	"fmt"

	"proyecto1/root/internal/config"
	"proyecto1/root/internal/messaging"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
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
	
	// Add credentials
	configOptions = append(configOptions, awsconfig.WithCredentialsProvider(
		credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		),
	))

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

// SendMessage sends a message to the SQS queue
func (s *SQSQueue) SendMessage(ctx context.Context, message messaging.Message) error {
	var messageBody string
	var err error

	// Handle different message types
	switch msg := message.(type) {
	case *messaging.VideoProcessingMessage:
		messageBodyBytes, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal video processing message: %w", err)
		}
		messageBody = string(messageBodyBytes)
	default:
		// For generic messages, use the body directly
		messageBody = message.GetBody()
	}

	_, err = s.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.queueURL),
		MessageBody: aws.String(messageBody),
	})
	if err != nil {
		return fmt.Errorf("failed to send message to SQS: %w", err)
	}

	return nil
}

// Close closes the connection to the message queue
func (s *SQSQueue) Close() error {
	// SQS doesn't require explicit connection closing
	return nil
}
