package httpserver

import (
	"log/slog"

	"mor80/service-reviewer/internal/handlers/pullrequest"
	"mor80/service-reviewer/internal/handlers/status"
	"mor80/service-reviewer/internal/handlers/team"
	"mor80/service-reviewer/internal/handlers/user"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	logger *slog.Logger,
	userHandler *user.UserHandler,
	teamHandler *team.TeamHandler,
	pullRequestHandler *pullrequest.PullRequestHandler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/ping", status.Ping)
	r.Head("/healthcheck", status.Healthcheck)

	userHandler.Register(r)
	teamHandler.Register(r)
	pullRequestHandler.Register(r)

	return r
}
