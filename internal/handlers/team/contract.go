package team

import (
	"context"

	"mor80/service-reviewer/internal/model"
)

type teamService interface {
	Create(ctx context.Context, team model.Team) (*model.Team, error)
	Get(ctx context.Context, teamName string) (*model.Team, error)
}
