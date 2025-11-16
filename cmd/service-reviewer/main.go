package main

import (
	"context"
	"log"
	"mor80/service-reviewer/internal/app"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := app.New(ctx, "./configs/default.yaml")
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		app.Shutdown(shutdownCtx)
	}()

	go func() {
		if err := app.Run(); err != nil {
			log.Printf("server stopped: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
}
