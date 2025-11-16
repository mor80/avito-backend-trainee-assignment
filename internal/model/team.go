package model

type Team struct {
	Name    string       `json:"team_name"`
	Members []TeamMember `json:"members"`
}

type TeamMember struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamDB struct {
	Name string `db:"team_name"`
}

type TeamDeactivationResult struct {
	TeamName          string   `json:"team_name"`
	DeactivatedUserID []string `json:"deactivated_user_ids"`
	ReassignedCount   int      `json:"reassigned_count"`
}
