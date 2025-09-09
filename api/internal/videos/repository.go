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
		INSERT INTO videos (title, status, is_public, user_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, status, is_public, uploaded_at, processed_at, deleted_at, user_id`

	var createdVideo Video
	err := r.db.QueryRow(query, video.Title, video.Status, video.IsPublic, video.UserID).Scan(
		&createdVideo.ID, &createdVideo.Title, &createdVideo.Status, &createdVideo.IsPublic,
		&createdVideo.UploadedAt, &createdVideo.ProcessedAt, 
		&createdVideo.DeletedAt, &createdVideo.UserID,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create video: %w", err)
	}

	return &createdVideo, nil
}

// GetVideoByID retrieves a video by its ID and user ID (ensures ownership)
func (r *Repository) GetVideoByID(videoID int, userID int) (*Video, error) {
	query := `
		SELECT id, title, status, is_public, uploaded_at, processed_at, deleted_at, user_id
		FROM videos 
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

	var video Video
	err := r.db.QueryRow(query, videoID, userID).Scan(
		&video.ID, &video.Title, &video.Status, &video.IsPublic,
		&video.UploadedAt, &video.ProcessedAt,
		&video.DeletedAt, &video.UserID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get video by ID %d for user %d: %w", videoID, userID, err)
	}

	return &video, nil
}

// GetVideosByUserID retrieves all videos for a specific user
func (r *Repository) GetVideosByUserID(userID int) ([]*Video, error) {
	query := `
		SELECT id, title, status, is_public, uploaded_at, processed_at, deleted_at, user_id
		FROM videos 
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY uploaded_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get videos for user %d: %w", userID, err)
	}
	defer rows.Close()

	var videos []*Video
	for rows.Next() {
		var video Video
		err := rows.Scan(
			&video.ID, &video.Title, &video.Status, &video.IsPublic,
			&video.UploadedAt, &video.ProcessedAt,
			&video.DeletedAt, &video.UserID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan video row: %w", err)
		}
		videos = append(videos, &video)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating video rows: %w", err)
	}

	return videos, nil
}

// SoftDeleteVideo marks a video as deleted by setting deleted_at timestamp
// Only allows deletion of private videos (is_public = false)
func (r *Repository) SoftDeleteVideo(videoID int, userID int) error {
	// First, check if video exists and get its details
	checkQuery := `
		SELECT is_public 
		FROM videos 
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
	
	var isPublic bool
	err := r.db.QueryRow(checkQuery, videoID, userID).Scan(&isPublic)
	if err != nil {
		return fmt.Errorf("video not found or not owned by user")
	}

	// Check if video is public (cannot be deleted)
	if isPublic {
		return fmt.Errorf("public videos cannot be deleted")
	}

	// Proceed with soft delete for private videos only
	deleteQuery := `
		UPDATE videos 
		SET deleted_at = NOW()
		WHERE id = $1 AND user_id = $2 AND is_public = false AND deleted_at IS NULL`

	result, err := r.db.Exec(deleteQuery, videoID, userID)
	if err != nil {
		return fmt.Errorf("failed to soft delete video: %w", err)
	}

	// Check if any rows were affected (should be 1 if successful)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("video deletion failed - video may be public or not found")
	}

	return nil
}

// GetPublicVideos retrieves all public videos that are not deleted
func (r *Repository) GetPublicVideos() ([]*Video, error) {
	query := `
		SELECT id, title, status, is_public, uploaded_at, processed_at, deleted_at, user_id
		FROM videos 
		WHERE is_public = true AND deleted_at IS NULL
		ORDER BY uploaded_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get public videos: %w", err)
	}
	defer rows.Close()

	var videos []*Video
	for rows.Next() {
		var video Video
		err := rows.Scan(
			&video.ID, &video.Title, &video.Status, &video.IsPublic,
			&video.UploadedAt, &video.ProcessedAt,
			&video.DeletedAt, &video.UserID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan public video row: %w", err)
		}
		videos = append(videos, &video)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating public video rows: %w", err)
	}

	return videos, nil
}
