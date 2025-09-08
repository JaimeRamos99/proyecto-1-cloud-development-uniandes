package videos

import (
	"fmt"

	"proyecto1/root/internal/database"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// CreateVideo creates a new video record in the database
func (r *Repository) CreateVideo(video *Video) (*Video, error) {
	query := `
		INSERT INTO videos (title, status, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, title, status, uploaded_at, processed_at, deleted_at, user_id`

	var createdVideo Video
	err := r.db.QueryRow(query, video.Title, video.Status, video.UserID).Scan(
		&createdVideo.ID, &createdVideo.Title, &createdVideo.Status,
		&createdVideo.UploadedAt, &createdVideo.ProcessedAt, 
		&createdVideo.DeletedAt, &createdVideo.UserID,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create video: %w", err)
	}

	return &createdVideo, nil
}
