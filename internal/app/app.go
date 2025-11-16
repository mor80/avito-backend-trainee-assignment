package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"mor80/service-reviewer/internal/config"
	"mor80/service-reviewer/internal/db/postgres"
	prhandler "mor80/service-reviewer/internal/handlers/pullrequest"
	teamhandler "mor80/service-reviewer/internal/handlers/team"
	userhandler "mor80/service-reviewer/internal/handlers/user"
	"mor80/service-reviewer/internal/httpserver"
	prrepo "mor80/service-reviewer/internal/repository/postgres/pullrequest"
	teamrepo "mor80/service-reviewer/internal/repository/postgres/team"
	userrepo "mor80/service-reviewer/internal/repository/postgres/user"
	prservice "mor80/service-reviewer/internal/service/pullrequest"
	teamservice "mor80/service-reviewer/internal/service/team"
	userservice "mor80/service-reviewer/internal/service/user"
	"mor80/service-reviewer/pkg/logger"
)

type App struct {
	config *config.Config
	logger *slog.Logger
	db     *pgxpool.Pool
	server *httpserver.Server
}

func New(ctx context.Context, configPath string) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("app: load config: %w", err)
	}

	log := logger.New(logger.EnvString(cfg.App.Env))

	pool, err := postgres.NewPool(ctx, cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("app: init postgres: %w", err)
	}

	userRepo := userrepo.New(pool)
	teamRepo := teamrepo.New(pool)
	pullRepo := prrepo.New(pool)

	userSvc := userservice.New(userRepo, pullRepo)
	pullSvc := prservice.New(pullRepo, userRepo, nil)
	teamSvc := teamservice.New(teamRepo, userRepo, pullRepo, pullSvc)

	userHandler := userhandler.New(userSvc)
	teamHandler := teamhandler.New(teamSvc)
	pullHandler := prhandler.New(pullSvc)

	router := httpserver.NewRouter(log, userHandler, teamHandler, pullHandler)
	server := httpserver.New(cfg.HTTP, log, router)

	return &App{
		config: cfg,
		logger: log,
		db:     pool,
		server: server,
	}, nil
}

func (a *App) Run() error {
	addr := fmt.Sprintf("%s:%d", a.config.HTTP.Host, a.config.HTTP.Port)
	a.logger.Info("service-reviewer starting", "env", a.config.App.Env, "addr", addr)

	return a.server.Start()
}

func (a *App) Shutdown(ctx context.Context) {
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("http server shutdown error", "err", err)
	}

	a.db.Close()
}
