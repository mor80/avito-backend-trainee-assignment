package user

import (
	"context"

	"mor80/service-reviewer/internal/model"
)

type userRepository interface {
	GetByID(ctx context.Context, userID string) (*model.User, error)
	ListByTeam(ctx context.Context, teamName string) ([]model.User, error)
	Upsert(ctx context.Context, users []model.User) error
	SetIsActive(ctx context.Context, userID string, isActive bool) (*model.User, error)
}
