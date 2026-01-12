package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"resume-tailor/internal/ai"
	"resume-tailor/internal/config"
	"resume-tailor/internal/db"
	"resume-tailor/internal/jobs"
	"resume-tailor/internal/resumes"
	"resume-tailor/internal/runreports"
	"resume-tailor/internal/runs"

	"github.com/google/uuid"
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
	runsRepoRaw := runs.NewRepo(pool)
	resumesRepo := resumes.NewRepo(pool)

	// Create adapter to avoid import cycle
	runsRepo := &runsRepoAdapter{repo: runsRepoRaw}

	// Initialize AI client (may be nil if API key is missing)
	var aiClient *ai.Client
	if cfg.OpenAIAPIKey != "" {
		var err error
		aiClient, err = ai.NewClientFromEnv(cfg.OpenAIAPIKey, cfg.OpenAIModel)
		if err != nil {
			slog.Error("failed to create AI client", "error", err)
			os.Exit(1)
		}
		slog.Info("AI client initialized", "model", cfg.OpenAIModel)
	} else {
		slog.Warn("OPENAI_API_KEY not set, worker will fail jobs that require AI")
	}

	worker := jobs.NewWorker(jobsRepo, pool, cfg.WorkerID, runreportsSvc, runsRepo, resumesRepo, aiClient)

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

// runsRepoAdapter adapts runs.Repo to jobs.RunsRepo interface
type runsRepoAdapter struct {
	repo *runs.Repo
}

func (a *runsRepoAdapter) GetRunByID(ctx context.Context, runID uuid.UUID) (jobs.RunData, error) {
	run, err := a.repo.GetRunByID(ctx, runID)
	if err != nil {
		return jobs.RunData{}, err
	}
	return jobs.RunData{
		ID:           run.ID,
		ResumeID:     run.ResumeID,
		JobText:      run.JobText,
		Status:       string(run.Status),
		ErrorMessage: run.ErrorMessage,
	}, nil
}
