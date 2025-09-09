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

// PlayerRankingResponse represents a single player in the rankings
type PlayerRankingResponse struct {
	UserID            int     `json:"user_id"`
	FirstName         string  `json:"first_name"`
	LastName          string  `json:"last_name"`
	Email             string  `json:"email"`
	City              string  `json:"city"`
	Country           string  `json:"country"`
	TotalVotes        int     `json:"total_votes"`
	Ranking           int     `json:"ranking"`
	LastUpdated       time.Time `json:"last_updated"`
}

// PlayerRankingsResponse represents the paginated response for rankings
type PlayerRankingsResponse struct {
	Rankings   []PlayerRankingResponse `json:"rankings"`
	Pagination PaginationResponse      `json:"pagination"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	CurrentPage int   `json:"current_page"`
	PageSize    int   `json:"page_size"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int   `json:"total_pages"`
}
