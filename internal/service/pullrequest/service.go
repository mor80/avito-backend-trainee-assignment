package pullrequest

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"mor80/service-reviewer/internal/model"
	"mor80/service-reviewer/internal/service"
)

const maxReviewers = 2

type random interface {
	Intn(n int) int
}

type PullRequestService struct {
	prRepo   service.PullRequestRepository
	userRepo service.UserRepository
	random   random
}

func New(prRepo service.PullRequestRepository, userRepo service.UserRepository, rng random) *PullRequestService {
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	return &PullRequestService{
		prRepo:   prRepo,
		userRepo: userRepo,
		random:   rng,
	}
}

func (s *PullRequestService) Create(ctx context.Context, pr model.PullRequest) (*model.PullRequest, error) {
	if err := validateCreateInput(pr); err != nil {
		return nil, fmt.Errorf("pull request service: %w", err)
	}

	author, err := s.userRepo.GetByID(ctx, pr.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("pull request service: %w", err)
	}

	teamMembers, err := s.userRepo.ListByTeam(ctx, author.TeamName)
	if err != nil {
		return nil, fmt.Errorf("pull request service: %w", err)
	}

	exclude := map[string]struct{}{author.ID: {}}
	reviewerIDs := s.selectReviewers(teamMembers, exclude, maxReviewers)

	now := time.Now().UTC()
	prDB := model.PullRequestDB{
		ID:        pr.ID,
		Name:      pr.Name,
		AuthorID:  pr.AuthorID,
		Status:    model.PullRequestStatusOpen,
		CreatedAt: &now,
		MergedAt:  nil,
	}

	created, err := s.prRepo.Create(ctx, prDB, reviewerIDs)
	if err != nil {
		return nil, fmt.Errorf("pull request service: %w", err)
	}

	return created, nil
}

func (s *PullRequestService) Merge(ctx context.Context, prID string) (*model.PullRequest, error) {
	if err := validatePullRequestID(prID); err != nil {
		return nil, fmt.Errorf("pull request service: %w", err)
	}

	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("pull request service: %w", err)
	}

	if pr.Status == model.PullRequestStatusMerged {
		return pr, nil
	}

	now := time.Now().UTC()
	updated, err := s.prRepo.UpdateStatus(ctx, prID, model.PullRequestStatusMerged, &now)
	if err != nil {
		return nil, fmt.Errorf("pull request service: %w", err)
	}

	return updated, nil
}

func (s *PullRequestService) Reassign(ctx context.Context, prID, oldReviewerID string) (*model.PullRequest, string, error) {
	if err := validatePullRequestID(prID); err != nil {
		return nil, "", fmt.Errorf("pull request service: %w", err)
	}

	if err := validateUserID(oldReviewerID, "old_user_id"); err != nil {
		return nil, "", fmt.Errorf("pull request service: %w", err)
	}

	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, "", fmt.Errorf("pull request service: %w", err)
	}

	if pr.Status == model.PullRequestStatusMerged {
		return nil, "", model.ErrPRMerged
	}

	if !containsReviewer(pr.AssignedReviewers, oldReviewerID) {
		return nil, "", model.ErrNotAssigned
	}

	oldReviewer, err := s.userRepo.GetByID(ctx, oldReviewerID)
	if err != nil {
		return nil, "", fmt.Errorf("pull request service: %w", err)
	}

	members, err := s.userRepo.ListByTeam(ctx, oldReviewer.TeamName)
	if err != nil {
		return nil, "", fmt.Errorf("pull request service: %w", err)
	}

	exclude := newReviewers(pr.AuthorID, oldReviewerID, pr.AssignedReviewers)
	candidates := filterMembers(members, exclude)
	if len(candidates) == 0 {
		return nil, "", model.ErrNoCandidate
	}

	replacement := candidates[s.random.Intn(len(candidates))]

	updated, err := s.prRepo.ReplaceReviewer(ctx, prID, oldReviewerID, replacement)
	if err != nil {
		return nil, "", fmt.Errorf("pull request service: %w", err)
	}

	return updated, replacement, nil
}

func (s *PullRequestService) AssignmentStats(ctx context.Context) ([]model.AssignmentStats, error) {
	stats, err := s.prRepo.GetAssignmentStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("pull request service: %w", err)
	}

	return stats, nil
}

func (s *PullRequestService) selectReviewers(members []model.User, exclude map[string]struct{}, limit int) []string {
	candidates := filterMembers(members, exclude)
	if len(candidates) <= limit {
		return candidates
	}

	return selectRandom(s.random, candidates, limit)
}

func validateCreateInput(pr model.PullRequest) error {
	if err := validatePullRequestID(pr.ID); err != nil {
		return err
	}

	if strings.TrimSpace(pr.Name) == "" {
		return fmt.Errorf("pull_request_name is required")
	}

	if err := validateUserID(pr.AuthorID, "author_id"); err != nil {
		return err
	}

	return nil
}

func validatePullRequestID(prID string) error {
	return notEmpty(prID, "pull_request_id")
}

func validateUserID(userID, field string) error {
	return notEmpty(userID, field)
}

func notEmpty(value, field string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}

	return nil
}

func containsReviewer(reviewers []string, id string) bool {
	for _, reviewerID := range reviewers {
		if reviewerID == id {
			return true
		}
	}

	return false
}

func newReviewers(authorID, oldReviewerID string, assigned []string) map[string]struct{} {
	reviewers := map[string]struct{}{
		authorID:      {},
		oldReviewerID: {},
	}

	for _, reviewerID := range assigned {
		reviewers[reviewerID] = struct{}{}
	}

	return reviewers
}

func filterMembers(members []model.User, exclude map[string]struct{}) []string {
	var ids []string

	for _, member := range members {
		if !member.IsActive {
			continue
		}

		if _, skip := exclude[member.ID]; skip {
			continue
		}

		ids = append(ids, member.ID)
	}

	return ids
}

func selectRandom(r random, ids []string, limit int) []string {
	if len(ids) <= limit {
		return append([]string(nil), ids...)
	}

	selected := make([]string, 0, limit)
	remaining := append([]string(nil), ids...)

	for len(selected) < limit && len(remaining) > 0 {
		idx := r.Intn(len(remaining))
		selected = append(selected, remaining[idx])
		remaining = append(remaining[:idx], remaining[idx+1:]...)
	}

	return selected
}
