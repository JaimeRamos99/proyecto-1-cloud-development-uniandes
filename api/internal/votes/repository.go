package votes

import (
	"database/sql"
	"fmt"

	"proyecto1/root/internal/database"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// VoteForVideo creates a new vote for a video by a user
func (r *Repository) VoteForVideo(userID, videoID int) error {
	query := `
		INSERT INTO votes (user_id, video_id)
		VALUES ($1, $2)
	`

	_, err := r.db.Exec(query, userID, videoID)
	if err != nil {
		// Check if it's a unique constraint violation (user already voted)
		if database.IsUniqueViolation(err) {
			return fmt.Errorf("user has already voted for this video")
		}
		return fmt.Errorf("failed to create vote: %w", err)
	}

	return nil
}

// RemoveVote removes a vote for a video by a user
func (r *Repository) RemoveVote(userID, videoID int) error {
	query := `
		DELETE FROM votes 
		WHERE user_id = $1 AND video_id = $2
	`

	result, err := r.db.Exec(query, userID, videoID)
	if err != nil {
		return fmt.Errorf("failed to remove vote: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vote not found")
	}

	return nil
}

// HasUserVoted checks if a user has already voted for a video
func (r *Repository) HasUserVoted(userID, videoID int) (bool, error) {
	query := `
		SELECT 1 FROM votes 
		WHERE user_id = $1 AND video_id = $2
		LIMIT 1
	`

	var exists int
	err := r.db.QueryRow(query, userID, videoID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check vote existence: %w", err)
	}

	return true, nil
}

// GetVideoVoteCount returns the total number of votes for a video
func (r *Repository) GetVideoVoteCount(videoID int) (int, error) {
	query := `
		SELECT COUNT(*) FROM votes 
		WHERE video_id = $1
	`

	var count int
	err := r.db.QueryRow(query, videoID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get vote count: %w", err)
	}

	return count, nil
}

// VideoExists checks if a video exists and is not soft-deleted
func (r *Repository) VideoExists(videoID int) (bool, error) {
	query := `
		SELECT 1 FROM videos 
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1
	`

	var exists int
	err := r.db.QueryRow(query, videoID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check video existence: %w", err)
	}

	return true, nil
}
