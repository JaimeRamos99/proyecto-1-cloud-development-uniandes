package votes

import (
	"time"
)

// Vote represents the vote model based on the database schema
type Vote struct {
	ID      int       `json:"id" db:"id"`
	UserID  int       `json:"user_id" db:"user_id"`
	VideoID int       `json:"video_id" db:"video_id"`
	VotedAt time.Time `json:"voted_at" db:"voted_at"`
}
