package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"resume-tailor/internal/config"
	"resume-tailor/internal/db"
	"resume-tailor/internal/jobs"
	"resume-tailor/internal/runreports"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to db", "error", err)
		os.Exit(1)
	}
	defer db.Close(pool)

	jobsRepo := jobs.NewRepo(pool)
	runreportsRepo := runreports.NewRepo(pool)
	runreportsSvc := runreports.NewService(runreportsRepo)
	worker := jobs.NewWorker(jobsRepo, pool, cfg.WorkerID, runreportsSvc)

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("shutdown signal received")
		cancel()
	}()

	// Run worker
	if err := worker.Run(ctx); err != nil {
		if err != context.Canceled {
			slog.Error("worker error", "error", err)
			os.Exit(1)
		}
	}

	slog.Info("worker stopped")
}
