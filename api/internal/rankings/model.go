package rankings

import (
	"time"
)

// PlayerRanking represents the player ranking model from the materialized view
type PlayerRanking struct {
	UserID            int       `json:"user_id" db:"user_id"`
	FirstName         string    `json:"first_name" db:"first_name"`
	LastName          string    `json:"last_name" db:"last_name"`
	Email             string    `json:"email" db:"email"`
	City              string    `json:"city" db:"city"`
	Country           string    `json:"country" db:"country"`
	TotalVotes        int       `json:"total_votes" db:"total_votes"`
	Ranking           int       `json:"ranking" db:"ranking"`
	LastUpdated       time.Time `json:"last_updated" db:"last_updated"`
}

// RankingFilters represents filters that can be applied to rankings
type RankingFilters struct {
	Country     string `form:"country"`
	City        string `form:"city"`
	MinVotes    *int   `form:"min_votes"`
	MaxVotes    *int   `form:"max_votes"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `form:"page,default=1" binding:"min=1"`
	PageSize int `form:"page_size,default=10" binding:"min=1,max=100"`
}

// GetOffset calculates the offset for database queries
func (p PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the page size
func (p PaginationParams) GetLimit() int {
	return p.PageSize
}
