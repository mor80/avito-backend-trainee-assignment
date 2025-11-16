package user

import (
	"context"

	"mor80/service-reviewer/internal/model"
)

type userService interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*model.User, error)
	GetReview(ctx context.Context, userID string) ([]model.PullRequestShort, error)
}
