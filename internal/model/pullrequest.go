package model

import (
	"fmt"
	"time"
)

type PullRequestStatus string

const (
	PullRequestStatusOpen   PullRequestStatus = "OPEN"
	PullRequestStatusMerged PullRequestStatus = "MERGED"
)

func (s PullRequestStatus) Valid() bool {
	switch s {
	case PullRequestStatusOpen, PullRequestStatusMerged:
		return true
	default:
		return false
	}
}

func (s PullRequestStatus) Validate() error {
	if s.Valid() {
		return nil
	}

	return fmt.Errorf("invalid pull request status: %s", s)
}

type PullRequest struct {
	ID                string            `json:"pull_request_id"`
	Name              string            `json:"pull_request_name"`
	AuthorID          string            `json:"author_id"`
	Status            PullRequestStatus `json:"status"`
	AssignedReviewers []string          `json:"assigned_reviewers"`
	CreatedAt         *time.Time        `json:"createdAt,omitempty"`
	MergedAt          *time.Time        `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	ID       string            `json:"pull_request_id"`
	Name     string            `json:"pull_request_name"`
	AuthorID string            `json:"author_id"`
	Status   PullRequestStatus `json:"status"`
}

type PullRequestDB struct {
	ID        string            `db:"pull_request_id"`
	Name      string            `db:"pull_request_name"`
	AuthorID  string            `db:"author_id"`
	Status    PullRequestStatus `db:"status"`
	CreatedAt *time.Time        `db:"created_at"`
	MergedAt  *time.Time        `db:"merged_at"`
}

type PullRequestReviewerDB struct {
	PullRequestID string `db:"pull_request_id"`
	ReviewerID    string `db:"reviewer_id"`
}

type AssignmentStats struct {
	UserID string `json:"user_id"`
	Count  int    `json:"assignment_count"`
}

type PullRequestAssignment struct {
	PullRequestID string
	ReviewerID    string
}
