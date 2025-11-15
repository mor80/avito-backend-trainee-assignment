package team

import (
	"context"

	"mor80/service-reviewer/internal/model"
)

type teamRepository interface {
	Create(ctx context.Context, teamName string) error
	Exists(ctx context.Context, teamName string) (bool, error)
	GetByName(ctx context.Context, teamName string) (*model.Team, error)
}
