package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"mor80/service-reviewer/internal/config"
)

const (
	defaultHost         = "0.0.0.0"
	defaultPort         = 8080
	defaultReadTimeout  = 5 * time.Second
	defaultWriteTimeout = 10 * time.Second
)

type Server struct {
	config config.HTTP
	logger *slog.Logger
	server *http.Server
}

func New(cfg config.HTTP, logger *slog.Logger, handler http.Handler) *Server {
	applyDefaults(&cfg)

	return &Server{
		config: cfg,
		logger: logger,
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
	}
}

func (s *Server) Start() error {
	s.logger.Info("HTTP server starting", "addr", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server failed to start: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	s.logger.Info("HTTP server shutting down")
	return s.server.Shutdown(shutdownCtx)
}

func applyDefaults(cfg *config.HTTP) {
	if cfg.Host == "" {
		cfg.Host = defaultHost
	}
	if cfg.Port == 0 {
		cfg.Port = defaultPort
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = defaultReadTimeout
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = defaultWriteTimeout
	}
}
