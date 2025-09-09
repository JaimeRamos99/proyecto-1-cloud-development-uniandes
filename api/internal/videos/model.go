package videos

import (
	"time"
)

// Video represents the video model based on the database schema
type Video struct {
	ID          int        `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Status      string     `json:"status" db:"status"`
	UploadedAt  time.Time  `json:"uploaded_at" db:"uploaded_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	UserID      int        `json:"user_id" db:"user_id"`
}

// VideoStatus constants
const (
	StatusUploaded   = "uploaded"
	StatusProcessed  = "processed"
)
