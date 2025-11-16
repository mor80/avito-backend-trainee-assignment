package team

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

type TeamHandler struct {
	service teamService
}

func New(service teamService) *TeamHandler {
	return &TeamHandler{service: service}
}

func (h *TeamHandler) Register(r chi.Router) {
	r.Post("/team/add", h.add)
	r.Get("/team/get", h.get)
}

func (h *TeamHandler) add(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req teamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, errorCodeBadRequest, "invalid request body")
		return
	}

	team := model.Team{
		Name:    req.TeamName,
		Members: toMembers(req.Members),
	}

	created, err := h.service.Create(r.Context(), team)
	if err != nil {
		status, code, msg := mapError(err)
		shared.WriteError(w, status, code, msg)
		return
	}

	shared.WriteJSON(w, http.StatusCreated, teamResponse{Team: created})
}

func (h *TeamHandler) get(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		shared.WriteError(w, http.StatusBadRequest, errorCodeBadRequest, "team_name is required")
		return
	}

	team, err := h.service.Get(r.Context(), teamName)
	if err != nil {
		status, code, msg := mapError(err)
		shared.WriteError(w, status, code, msg)
		return
	}

	shared.WriteJSON(w, http.StatusOK, team)
}

func toMembers(items []teamMemberObject) []model.TeamMember {
	members := make([]model.TeamMember, len(items))

	for i, item := range items {
		members[i] = model.TeamMember{
			ID:       item.UserID,
			Username: item.Username,
			IsActive: item.IsActive,
		}
	}

	return members
}

func mapError(err error) (int, string, string) {
	var domainErr model.DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case model.ErrorCodeNotFound:
			return http.StatusNotFound, string(domainErr.Code), domainErr.Message
		case model.ErrorCodeTeamExists:
			return http.StatusBadRequest, string(domainErr.Code), domainErr.Message
		default:
			return http.StatusBadRequest, string(domainErr.Code), domainErr.Message
		}
	}

	return http.StatusInternalServerError, errorCodeInternal, "internal server error"
}
