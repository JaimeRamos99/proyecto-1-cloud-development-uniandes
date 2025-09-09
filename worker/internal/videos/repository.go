package videos

import (
	"fmt"

	"worker/internal/database"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// UpdateVideoStatus updates the status of a video record
func (r *Repository) UpdateVideoStatus(videoID int, status string) error {
	var query string
	var args []interface{}
	
	if status == StatusProcessed {
		// Set processed_at timestamp when status is processed
		query = `
			UPDATE videos 
			SET status = $1, processed_at = CURRENT_TIMESTAMP
			WHERE id = $2`
		args = []interface{}{status, videoID}
	} else {
		// Update status without setting processed_at
		query = `
			UPDATE videos 
			SET status = $1
			WHERE id = $2`
		args = []interface{}{status, videoID}
	}
	
	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update video status: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("video with ID %d not found", videoID)
	}
	
	return nil
}

// GetVideoByS3Key retrieves a video by its S3 key
func (r *Repository) GetVideoByS3Key(s3Key string) (*Video, error) {
	query := `
		SELECT id, title, status, uploaded_at, processed_at, deleted_at, user_id
		FROM videos
		WHERE id = (
			SELECT CAST(SUBSTRING($1 FROM '^([0-9]+)\.') AS INTEGER)
		)`
	
	var video Video
	err := r.db.QueryRow(query, s3Key).Scan(
		&video.ID, &video.Title, &video.Status,
		&video.UploadedAt, &video.ProcessedAt,
		&video.DeletedAt, &video.UserID,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get video by S3 key: %w", err)
	}
	
	return &video, nil
}
