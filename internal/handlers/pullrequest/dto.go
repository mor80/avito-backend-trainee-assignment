package pullrequest

import "mor80/service-reviewer/internal/model"

type createRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type mergeRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type reassignRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

type prResponse struct {
	PR *model.PullRequest `json:"pr"`
}

type reassignResponse struct {
	PR         *model.PullRequest `json:"pr"`
	ReplacedBy string             `json:"replaced_by"`
}
