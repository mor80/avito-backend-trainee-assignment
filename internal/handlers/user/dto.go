package user

import "mor80/service-reviewer/internal/model"

type setIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type setIsActiveResponse struct {
	User *model.User `json:"user"`
}

type getReviewResponse struct {
	UserID       string                   `json:"user_id"`
	PullRequests []model.PullRequestShort `json:"pull_requests"`
}

