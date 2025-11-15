package pullrequest

import (
	"context"
	"time"

	"mor80/service-reviewer/internal/model"
)

type pullRequestRepository interface {
	Create(ctx context.Context, pr model.PullRequestDB, reviewerIDs []string) (*model.PullRequest, error)
	GetByID(ctx context.Context, prID string) (*model.PullRequest, error)
	UpdateStatus(ctx context.Context, prID string, status model.PullRequestStatus, mergedAt *time.Time) (*model.PullRequest, error)
	ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) (*model.PullRequest, error)
	ListByReviewer(ctx context.Context, reviewerID string) ([]model.PullRequestShort, error)
}
