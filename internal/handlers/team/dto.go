package team

import "mor80/service-reviewer/internal/model"

type teamRequest struct {
	TeamName string             `json:"team_name"`
	Members  []teamMemberObject `json:"members"`
}

type teamResponse struct {
	Team *model.Team `json:"team"`
}

type teamMemberObject struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

