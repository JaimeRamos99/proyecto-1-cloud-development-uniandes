package dto

import "time"

// VoteResponse represents the response after voting for a video
type VoteResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	VideoID   int       `json:"video_id"`
	UserID    int       `json:"user_id"`
	VotedAt   time.Time `json:"voted_at"`
	VoteCount int       `json:"vote_count"`
}

// UnvoteResponse represents the response after removing a vote
type UnvoteResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	VideoID   int    `json:"video_id"`
	UserID    int    `json:"user_id"`
	VoteCount int    `json:"vote_count"`
}
