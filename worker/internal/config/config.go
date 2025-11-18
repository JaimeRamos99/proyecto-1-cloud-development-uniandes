package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the worker application
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	AWS      AWSConfig
	Retry    RetryConfig
}

type AppConfig struct {
	Name    string
	Version string
	Env     string // development, staging, production
}

type DatabaseConfig struct {
	Host         string
	Port         string
	Name         string
	User         string
	Password     string
	SSLMode      string
	Driver       string // postgres, memory
	MaxOpenConns int
	MaxIdleConns int
}

type AWSConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	EndpointURL     string // For LocalStack
	S3BucketName    string
	SQSQueueName    string
	DLQQueueName    string // Dead Letter Queue name
}

// RetryConfig holds retry policy configuration
type RetryConfig struct {
	MaxRetries    int  // Maximum number of retry attempts
	BaseDelay     int  // Base delay in seconds for exponential backoff
	MaxDelay      int  // Maximum delay in seconds
	EnableBackoff bool // Enable exponential backoff
}

// Load reads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		App: AppConfig{
			Name:    getEnv("APP_NAME", "Proyecto_1_Worker"),
			Version: getEnv("APP_VERSION", "1.0.0"),
			Env:     getEnv("APP_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "5432"),
			Name:         getEnv("DB_NAME", "proyecto_1"),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", "password"),
			SSLMode:      getEnv("DB_SSL_MODE", "disable"),
			Driver:       getEnv("DB_DRIVER", "postgres"), // only postgres supported
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 5),
		},
	AWS: AWSConfig{
		// DO NOT read AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY from env
		// Lambda execution role credentials will be used automatically
		AccessKeyID:     "", // Leave empty to use IAM role
		SecretAccessKey: "", // Leave empty to use IAM role
		Region:          getEnv("AWS_REGION", "us-east-1"),
		EndpointURL:     getEnv("AWS_ENDPOINT_URL", ""), // Empty for real AWS
		S3BucketName:    getEnv("S3_BUCKET_NAME", "proyecto1-videos"),
		SQSQueueName:    getEnv("SQS_QUEUE_NAME", "proyecto1-video-processing"),
		DLQQueueName:    getEnv("DLQ_QUEUE_NAME", "proyecto1-video-processing-dlq"),
	},
		Retry: RetryConfig{
			MaxRetries:    getEnvInt("WORKER_MAX_RETRIES", 3),
			BaseDelay:     getEnvInt("WORKER_BASE_DELAY", 2), // 2 seconds base delay
			MaxDelay:      getEnvInt("WORKER_MAX_DELAY", 60), // 60 seconds max delay
			EnableBackoff: getEnvBool("WORKER_ENABLE_BACKOFF", true),
		},
	}
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as integer with a fallback default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool gets an environment variable as boolean with a fallback default
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
