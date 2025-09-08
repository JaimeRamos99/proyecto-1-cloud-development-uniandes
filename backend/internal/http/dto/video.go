package dto

import "time"

// VideoUploadResponse represents the response for successful video upload
type VideoUploadResponse struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	UploadedAt time.Time `json:"uploaded_at"`
	UserID     int       `json:"user_id"`
	S3Key      string    `json:"s3_key,omitempty"` // S3 storage key
}
