package dto

import "time"

// VideoUploadResponse represents the response for successful video upload
type VideoUploadResponse struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	IsPublic   bool      `json:"is_public"`
	UploadedAt time.Time `json:"uploaded_at"`
	UserID     int       `json:"user_id"`
	S3Key      string    `json:"s3_key,omitempty"` // S3 storage key
}

// VideoResponse represents the response for video details
type VideoResponse struct {
	VideoID      int        `json:"video_id"`
	Title        string     `json:"title"`
	Status       string     `json:"status"`
	IsPublic     bool       `json:"is_public"`
	UploadedAt   time.Time  `json:"uploaded_at"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
	OriginalURL  string     `json:"original_url"`
	ProcessedURL string     `json:"processed_url"`
	Votes        int        `json:"votes"`
}

// PublicVideoResponse represents the response for public video details (without original URL)
type PublicVideoResponse struct {
	VideoID      int        `json:"video_id"`
	Title        string     `json:"title"`
	Status       string     `json:"status"`
	IsPublic     bool       `json:"is_public"`
	UploadedAt   time.Time  `json:"uploaded_at"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
	ProcessedURL string     `json:"processed_url"`
	Votes        int        `json:"votes"`
}
