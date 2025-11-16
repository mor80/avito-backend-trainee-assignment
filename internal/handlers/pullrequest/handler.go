package pullrequest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"mor80/service-reviewer/internal/handlers/shared"
	"mor80/service-reviewer/internal/model"
)

const (
	errorCodeBadRequest = "BAD_REQUEST"
	errorCodeInternal   = "INTERNAL_ERROR"
)

type PullRequestHandler struct {
	service pullRequestService
}

func New(service pullRequestService) *PullRequestHandler {
	return &PullRequestHandler{service: service}
}

func (h *PullRequestHandler) Register(r chi.Router) {
	r.Post("/pullRequest/create", h.create)
	r.Post("/pullRequest/merge", h.merge)
	r.Post("/pullRequest/reassign", h.reassign)
}

func (h *PullRequestHandler) create(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, errorCodeBadRequest, "invalid request body")
		return
	}

	if req.PullRequestID == "" || req.PullRequestName == "" || req.AuthorID == "" {
		shared.WriteError(w, http.StatusBadRequest, errorCodeBadRequest, "pull_request_id, pull_request_name and author_id are required")
		return
	}

	pr := model.PullRequest{
		ID:       req.PullRequestID,
		Name:     req.PullRequestName,
		AuthorID: req.AuthorID,
	}

	created, err := h.service.Create(r.Context(), pr)
	if err != nil {
		status, code, msg := mapError(err)
		shared.WriteError(w, status, code, msg)
		return
	}

	shared.WriteJSON(w, http.StatusCreated, prResponse{PR: created})
}

func (h *PullRequestHandler) merge(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req mergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, errorCodeBadRequest, "invalid request body")
		return
	}

	if req.PullRequestID == "" {
		shared.WriteError(w, http.StatusBadRequest, errorCodeBadRequest, "pull_request_id is required")
		return
	}

	pr, err := h.service.Merge(r.Context(), req.PullRequestID)
	if err != nil {
		status, code, msg := mapError(err)
		shared.WriteError(w, status, code, msg)
		return
	}

	shared.WriteJSON(w, http.StatusOK, prResponse{PR: pr})
}

func (h *PullRequestHandler) reassign(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req reassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, errorCodeBadRequest, "invalid request body")
		return
	}

	if req.PullRequestID == "" || req.OldUserID == "" {
		shared.WriteError(w, http.StatusBadRequest, errorCodeBadRequest, "pull_request_id and old_user_id are required")
		return
	}

	pr, replacedBy, err := h.service.Reassign(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		status, code, msg := mapError(err)
		shared.WriteError(w, status, code, msg)
		return
	}

	shared.WriteJSON(w, http.StatusOK, reassignResponse{
		PR:         pr,
		ReplacedBy: replacedBy,
	})
}

func mapError(err error) (int, string, string) {
	var domainErr model.DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case model.ErrorCodeNotFound:
			return http.StatusNotFound, string(domainErr.Code), domainErr.Message
		case model.ErrorCodePRExists,
			model.ErrorCodePRMerged,
			model.ErrorCodeNotAssigned,
			model.ErrorCodeNoCandidate:
			return http.StatusConflict, string(domainErr.Code), domainErr.Message
		default:
			return http.StatusBadRequest, string(domainErr.Code), domainErr.Message
		}
	}

	return http.StatusInternalServerError, errorCodeInternal, "internal server error"
}
