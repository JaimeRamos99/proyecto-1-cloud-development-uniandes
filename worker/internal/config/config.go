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
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			Region:          getEnv("AWS_DEFAULT_REGION", "us-east-1"),
			EndpointURL:     getEnv("AWS_ENDPOINT_URL", ""), // Empty for real AWS
			S3BucketName:    getEnv("S3_BUCKET_NAME", "proyecto1-videos"),
			SQSQueueName:    getEnv("SQS_QUEUE_NAME", "proyecto1-video-processing"),
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
