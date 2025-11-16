package service

import (
	"context"
	"time"

	"mor80/service-reviewer/internal/model"
)

type UserRepository interface {
	GetByID(ctx context.Context, userID string) (*model.User, error)
	ListByTeam(ctx context.Context, teamName string) ([]model.User, error)
	Upsert(ctx context.Context, users []model.User) error
	SetIsActive(ctx context.Context, userID string, isActive bool) (*model.User, error)
	ListByIDs(ctx context.Context, teamName string, userIDs []string) ([]model.User, error)
	DeactivateUsers(ctx context.Context, teamName string, userIDs []string) ([]string, error)
}

type TeamRepository interface {
	Create(ctx context.Context, teamName string) error
	Exists(ctx context.Context, teamName string) (bool, error)
	GetByName(ctx context.Context, teamName string) (*model.Team, error)
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr model.PullRequestDB, reviewerIDs []string) (*model.PullRequest, error)
	GetByID(ctx context.Context, prID string) (*model.PullRequest, error)
	UpdateStatus(ctx context.Context, prID string, status model.PullRequestStatus, mergedAt *time.Time) (*model.PullRequest, error)
	ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) (*model.PullRequest, error)
	ListByReviewer(ctx context.Context, reviewerID string) ([]model.PullRequestShort, error)
	ListOpenAssignmentsByReviewers(ctx context.Context, reviewerIDs []string) ([]model.PullRequestAssignment, error)
	GetAssignmentStats(ctx context.Context) ([]model.AssignmentStats, error)
}
