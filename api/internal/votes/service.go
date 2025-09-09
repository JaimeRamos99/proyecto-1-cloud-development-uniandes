package votes

import (
	"fmt"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

// VoteForVideo allows a user to vote for a video
func (s *Service) VoteForVideo(userID, videoID int) error {
	// Check if video exists and is not deleted
	exists, err := s.repository.VideoExists(videoID)
	if err != nil {
		return fmt.Errorf("failed to verify video existence: %w", err)
	}
	
	if !exists {
		return fmt.Errorf("video not found or has been deleted")
	}
	
	// Check if user has already voted for this video
	hasVoted, err := s.repository.HasUserVoted(userID, videoID)
	if err != nil {
		return fmt.Errorf("failed to check existing vote: %w", err)
	}
	
	if hasVoted {
		return fmt.Errorf("user has already voted for this video")
	}
	
	// Create the vote
	err = s.repository.VoteForVideo(userID, videoID)
	if err != nil {
		return fmt.Errorf("failed to cast vote: %w", err)
	}
	
	return nil
}

// RemoveVote allows a user to remove their vote from a video
func (s *Service) RemoveVote(userID, videoID int) error {
	// Check if video exists and is not deleted
	exists, err := s.repository.VideoExists(videoID)
	if err != nil {
		return fmt.Errorf("failed to verify video existence: %w", err)
	}
	
	if !exists {
		return fmt.Errorf("video not found or has been deleted")
	}
	
	// Remove the vote
	err = s.repository.RemoveVote(userID, videoID)
	if err != nil {
		return fmt.Errorf("failed to remove vote: %w", err)
	}
	
	return nil
}

// GetVideoVoteCount returns the total number of votes for a video
func (s *Service) GetVideoVoteCount(videoID int) (int, error) {
	count, err := s.repository.GetVideoVoteCount(videoID)
	if err != nil {
		return 0, fmt.Errorf("failed to get vote count: %w", err)
	}
	
	return count, nil
}

// HasUserVoted checks if a user has voted for a specific video
func (s *Service) HasUserVoted(userID, videoID int) (bool, error) {
	hasVoted, err := s.repository.HasUserVoted(userID, videoID)
	if err != nil {
		return false, fmt.Errorf("failed to check vote status: %w", err)
	}
	
	return hasVoted, nil
}
