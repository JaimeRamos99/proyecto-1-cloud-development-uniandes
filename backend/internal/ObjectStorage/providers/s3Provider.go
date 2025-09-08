package providers

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Config holds the configuration for S3 provider
type S3Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	BucketName      string
	EndpointURL     string // For LocalStack
}

// S3Provider implements IFileStorageProvider using AWS S3
type S3Provider struct {
	client     *s3.Client
	bucketName string
}

// NewS3Provider creates a new S3 provider instance
func NewS3Provider(cfg *S3Config) (*S3Provider, error) {
	// Load base AWS configuration
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with custom options
	var client *s3.Client
	if cfg.EndpointURL != "" {
		// LocalStack configuration - use custom endpoint
		client = s3.NewFromConfig(awsConfig, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.EndpointURL)
			o.UsePathStyle = true // Required for LocalStack
		})
	} else {
		// Real AWS configuration
		client = s3.NewFromConfig(awsConfig)
	}

	return &S3Provider{
		client:     client,
		bucketName: cfg.BucketName,
	}, nil
}

// UploadFile uploads a file buffer to S3
func (s *S3Provider) UploadFile(fileBuffer []byte, fileName string) error {
	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fileName),
		Body:   bytes.NewReader(fileBuffer),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}
	return nil
}

// GetSignedUrl generates a presigned URL for file access
func (s *S3Provider) GetSignedUrl(fileName string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	
	request, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fileName),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute // Default 15 minutes
	})
	
	if err != nil {
		return "", fmt.Errorf("failed to create presigned URL: %w", err)
	}
	
	return request.URL, nil
}

// DeleteFile deletes a file from S3
func (s *S3Provider) DeleteFile(fileName string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}
	return nil
}

// IFileStorageProvider defines the interface for file storage operations
type IFileStorageProvider interface {
	UploadFile(fileBuffer []byte, fileName string) error
	GetSignedUrl(fileName string) (string, error)
	DeleteFile(fileName string) error
}

// Ensure S3Provider implements IFileStorageProvider
var _ IFileStorageProvider = (*S3Provider)(nil)
