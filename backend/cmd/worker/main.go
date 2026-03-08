package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"time"

	"resume-tailor/internal/ai"
	"resume-tailor/internal/artifacts"
	"resume-tailor/internal/auth"
	"resume-tailor/internal/config"
	"resume-tailor/internal/db"
	"resume-tailor/internal/jobs"
	"resume-tailor/internal/monitoring"
	"resume-tailor/internal/resumes"
	"resume-tailor/internal/runreports"
	"resume-tailor/internal/runs"
	"resume-tailor/migrations"

	"github.com/google/uuid"
)

func main() {
	ctx := context.Background()

	// Initialize monitoring (Sentry when SENTRY_DSN is set).
	monitoring.Init("resume-tailor-worker")
	defer monitoring.Flush(2 * time.Second)

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

	slog.Info("running migrations")
	if err := migrations.Run(ctx, pool); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	jobsRepo := jobs.NewRepo(pool)
	runreportsRepo := runreports.NewRepo(pool)
	runreportsSvc := runreports.NewService(runreportsRepo)
	runsRepoRaw := runs.NewRepo(pool)
	resumesRepo := resumes.NewRepo(pool)
	artifactsRepo := artifacts.NewRepo(pool)
	artifactsSvc := artifacts.NewService(artifactsRepo)
	authRepo := auth.NewRepo(pool)

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

	worker := jobs.NewWorker(
		jobsRepo,
		pool,
		cfg.WorkerID,
		runreportsSvc,
		runsRepo,
		resumesRepo,
		aiClient,
		artifactsSvc,
		cfg.PDFEnabled,
		cfg.TectonicBin,
		cfg.WorkerJobTimeout,
		cfg.DisciplineMode,
		authRepo,
		cfg.APIKeyEncryptionSecret,
		cfg.OpenAIModel,
	)

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
	var discipline *string
	if run.Discipline != nil {
		value := string(*run.Discipline)
		discipline = &value
	}
	return jobs.RunData{
		ID:               run.ID,
		UserID:           run.UserID,
		ResumeID:         run.ResumeID,
		JobText:          run.JobText,
		ProjectControls:  mapProjectControls(run.ProjectControls),
		Discipline:       discipline,
		DisciplineScore:  run.DisciplineScore,
		DisciplineSource: string(run.DisciplineSource),
		Status:           string(run.Status),
		ErrorMessage:     run.ErrorMessage,
	}, nil
}

func (a *runsRepoAdapter) UpdateRunDiscipline(ctx context.Context, runID uuid.UUID, discipline string, confidence float64, source string) error {
	parsedDiscipline, ok := runs.ParseDiscipline(discipline)
	if !ok {
		return runs.ErrBadInput
	}
	parsedSource := runs.DisciplineSource(source)
	if parsedSource != runs.DisciplineSourceAuto && parsedSource != runs.DisciplineSourceUserOverride {
		return runs.ErrBadInput
	}
	return a.repo.UpdateRunDiscipline(ctx, runID, parsedDiscipline, confidence, parsedSource)
}

func mapProjectControls(controls []runs.ProjectControl) []ai.ProjectControl {
	if len(controls) == 0 {
		return nil
	}
	out := make([]ai.ProjectControl, 0, len(controls))
	for _, c := range controls {
		out = append(out, ai.ProjectControl{
			Name: c.Name,
			Mode: string(c.Mode),
		})
	}
	return out
}
