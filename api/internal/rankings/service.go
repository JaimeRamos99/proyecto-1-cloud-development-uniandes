package rankings

import (
	"math"
	"proyecto1/root/internal/http/dto"
)

type Service struct {
	repo *Repository
}

// NewService creates a new rankings service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetPlayerRankings retrieves player rankings with pagination and filters
func (s *Service) GetPlayerRankings(filters RankingFilters, pagination PaginationParams) (*dto.PlayerRankingsResponse, error) {
	// Get rankings from repository
	rankings, totalCount, err := s.repo.GetPlayerRankings(filters, pagination)
	if err != nil {
		return nil, err
	}

	// Convert to DTOs
	var rankingDTOs []dto.PlayerRankingResponse
	for _, ranking := range rankings {
		rankingDTOs = append(rankingDTOs, dto.PlayerRankingResponse{
			UserID:      ranking.UserID,
			FirstName:   ranking.FirstName,
			LastName:    ranking.LastName,
			Email:       ranking.Email,
			City:        ranking.City,
			Country:     ranking.Country,
			TotalVotes:  ranking.TotalVotes,
			Ranking:     ranking.Ranking,
			LastUpdated: ranking.LastUpdated,
		})
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(totalCount) / float64(pagination.PageSize)))

	paginationResponse := dto.PaginationResponse{
		CurrentPage: pagination.Page,
		PageSize:    pagination.PageSize,
		TotalItems:  totalCount,
		TotalPages:  totalPages,
	}

	return &dto.PlayerRankingsResponse{
		Rankings:   rankingDTOs,
		Pagination: paginationResponse,
	}, nil
}

// RefreshRankings manually refreshes the rankings
func (s *Service) RefreshRankings() error {
	return s.repo.RefreshRankings()
}
