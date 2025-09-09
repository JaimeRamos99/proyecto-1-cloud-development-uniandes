package rankings

import (
	"fmt"
	"proyecto1/root/internal/database"
	"strings"
)

type Repository struct {
	db *database.DB
}

// NewRepository creates a new rankings repository
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// GetPlayerRankings retrieves player rankings with pagination and filters
func (r *Repository) GetPlayerRankings(filters RankingFilters, pagination PaginationParams) ([]PlayerRanking, int64, error) {
	// Build WHERE clause based on filters
	var whereClauses []string
	var args []interface{}
	argIndex := 1

	if filters.Country != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("LOWER(country) = LOWER($%d)", argIndex))
		args = append(args, filters.Country)
		argIndex++
	}

	if filters.City != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("LOWER(city) = LOWER($%d)", argIndex))
		args = append(args, filters.City)
		argIndex++
	}

	if filters.MinVotes != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("total_votes >= $%d", argIndex))
		args = append(args, *filters.MinVotes)
		argIndex++
	}

	if filters.MaxVotes != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("total_votes <= $%d", argIndex))
		args = append(args, *filters.MaxVotes)
		argIndex++
	}

	var whereClause string
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// First, get the total count for pagination
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM player_rankings 
		%s
	`, whereClause)

	var totalCount int64
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Now get the actual rankings with pagination
	query := fmt.Sprintf(`
		SELECT 
			user_id, first_name, last_name, email, city, country,
			total_votes, ranking, last_updated
		FROM player_rankings 
		%s
		ORDER BY ranking ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	// Add pagination parameters to args
	args = append(args, pagination.GetLimit(), pagination.GetOffset())

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query rankings: %w", err)
	}
	defer rows.Close()

	var rankings []PlayerRanking
	for rows.Next() {
		var ranking PlayerRanking
		err := rows.Scan(
			&ranking.UserID,
			&ranking.FirstName,
			&ranking.LastName,
			&ranking.Email,
			&ranking.City,
			&ranking.Country,
			&ranking.TotalVotes,
			&ranking.Ranking,
			&ranking.LastUpdated,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan ranking: %w", err)
		}
		rankings = append(rankings, ranking)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating rows: %w", err)
	}

	return rankings, totalCount, nil
}

// RefreshRankings manually refreshes the materialized view
func (r *Repository) RefreshRankings() error {
	_, err := r.db.Exec("SELECT refresh_player_rankings()")
	if err != nil {
		return fmt.Errorf("failed to refresh rankings: %w", err)
	}
	return nil
}

