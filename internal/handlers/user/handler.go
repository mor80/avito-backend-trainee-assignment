package user

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
	errorInvalidJSON    = "invalid request body"
	errorMissingUserID  = "user_id is required"
)

type UserHandler struct {
	service userService
}

func New(service userService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Register(r chi.Router) {
	r.Post("/users/setIsActive", h.setIsActive)
	r.Get("/users/getReview", h.getReview)
}

func (h *UserHandler) setIsActive(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req setIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, errorCodeBadRequest, errorInvalidJSON)
		return
	}

	if req.UserID == "" {
		shared.WriteError(w, http.StatusBadRequest, errorCodeBadRequest, errorMissingUserID)
		return
	}

	user, err := h.service.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		status, code, msg := mapError(err)
		shared.WriteError(w, status, code, msg)
		return
	}

	shared.WriteJSON(w, http.StatusOK, setIsActiveResponse{User: user})
}

func (h *UserHandler) getReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		shared.WriteError(w, http.StatusBadRequest, errorCodeBadRequest, errorMissingUserID)
		return
	}

	prs, err := h.service.GetReview(r.Context(), userID)
	if err != nil {
		status, code, msg := mapError(err)
		shared.WriteError(w, status, code, msg)
		return
	}

	shared.WriteJSON(w, http.StatusOK, getReviewResponse{
		UserID:       userID,
		PullRequests: prs,
	})
}

func mapError(err error) (int, string, string) {
	var domainErr model.DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case model.ErrorCodeNotFound:
			return http.StatusNotFound, string(domainErr.Code), domainErr.Message
		case model.ErrorCodeTeamExists,
			model.ErrorCodePRExists,
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
