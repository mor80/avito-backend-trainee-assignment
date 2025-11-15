package logger

import (
	"log/slog"
	"os"
)

type EnvString string

const (
	envLocal EnvString = "local"
	envDev   EnvString = "dev"
	envProd  EnvString = "prod"
)

func New(env EnvString) *slog.Logger {
	var l *slog.Logger

	switch env {
	case envLocal:
		l = slog.New(
			slog.NewTextHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelDebug},
			),
		)
	case envDev:
		l = slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelDebug},
			),
		)
	case envProd:
		l = slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelInfo},
			),
		)
	default:
		l = slog.New(
			slog.NewTextHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelInfo},
			),
		)
	}

	return l
}
