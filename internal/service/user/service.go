package user

import (
	"context"
	"fmt"
	"strings"

	"mor80/service-reviewer/internal/model"
	"mor80/service-reviewer/internal/service"
)

type UserService struct {
	userRepo service.UserRepository
	prRepo   service.PullRequestRepository
}

func New(userRepo service.UserRepository, prRepo service.PullRequestRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*model.User, error) {
	if err := validateUserID(userID); err != nil {
		return nil, fmt.Errorf("user service: %w", err)
	}

	user, err := s.userRepo.SetIsActive(ctx, userID, isActive)
	if err != nil {
		return nil, fmt.Errorf("user service: %w", err)
	}

	return user, nil
}

func (s *UserService) GetReview(ctx context.Context, userID string) ([]model.PullRequestShort, error) {
	if err := validateUserID(userID); err != nil {
		return nil, fmt.Errorf("user service: %w", err)
	}

	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return nil, fmt.Errorf("user service: %w", err)
	}

	prs, err := s.prRepo.ListByReviewer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user service: %w", err)
	}

	return prs, nil
}

func validateUserID(userID string) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("user_id is required")
	}

	return nil
}
