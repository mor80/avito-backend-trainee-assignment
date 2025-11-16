package pullrequest

import (
	"context"

	"mor80/service-reviewer/internal/model"
)

type pullRequestService interface {
	Create(ctx context.Context, pr model.PullRequest) (*model.PullRequest, error)
	Merge(ctx context.Context, prID string) (*model.PullRequest, error)
	Reassign(ctx context.Context, prID, oldReviewerID string) (*model.PullRequest, string, error)
	AssignmentStats(ctx context.Context) ([]model.AssignmentStats, error)
}
