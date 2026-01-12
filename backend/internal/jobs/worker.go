package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"resume-tailor/internal/ai"
	"resume-tailor/internal/resumes"
	"resume-tailor/internal/runreports"
	"resume-tailor/internal/scoring/bm25"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const pollInterval = 1 * time.Second

const (
	runStatusCreated    = "created"
	runStatusQueued     = "queued"
	runStatusProcessing = "processing"
	runStatusFailed     = "failed"
	runStatusCompleted  = "completed"
)

// RunsRepo is an interface to avoid import cycle with runs package
type RunsRepo interface {
	GetRunByID(ctx context.Context, runID uuid.UUID) (RunData, error)
}

// RunData represents the run data needed by the worker
type RunData struct {
	ID           uuid.UUID
	ResumeID     uuid.UUID
	JobText      string
	Status       string
	ErrorMessage *string
}

type Worker struct {
	jobsRepo    *Repo
	db          *pgxpool.Pool
	workerID    string
	reportsSvc  *runreports.Service
	runsRepo    RunsRepo
	resumesRepo *resumes.Repo
	aiClient    *ai.Client
}

func NewWorker(jobsRepo *Repo, db *pgxpool.Pool, workerID string, reportsSvc *runreports.Service, runsRepo RunsRepo, resumesRepo *resumes.Repo, aiClient *ai.Client) *Worker {
	return &Worker{
		jobsRepo:    jobsRepo,
		db:          db,
		workerID:    workerID,
		reportsSvc:  reportsSvc,
		runsRepo:    runsRepo,
		resumesRepo: resumesRepo,
		aiClient:    aiClient,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	slog.Info("worker started", "worker_id", w.workerID)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("worker stopping", "worker_id", w.workerID)
			return ctx.Err()
		case <-ticker.C:
			if err := w.processNextJob(ctx); err != nil {
				if err == ErrNoJobs {
					// No jobs available, continue polling
					continue
				}
				slog.Error("error processing job", "error", err, "worker_id", w.workerID)
			}
		}
	}
}

func (w *Worker) processNextJob(ctx context.Context) error {
	// Claim next job
	job, err := w.jobsRepo.ClaimNextProcessRun(ctx, w.workerID)
	if err != nil {
		return err
	}

	slog.Info("claimed job", "job_id", job.ID, "run_id", job.RunID, "worker_id", w.workerID)

	// Update run status to processing
	if err := w.updateRunStatus(ctx, job.RunID, runStatusProcessing, nil); err != nil {
		slog.Error("failed to update run status to processing", "error", err, "run_id", job.RunID)
		// Mark job as failed
		w.jobsRepo.MarkJobFailed(ctx, job.ID, fmt.Sprintf("failed to update run status: %v", err), job.Attempts < job.MaxAttempts)
		return err
	}

	// Process the run (MVP stub)
	if err := w.processRun(ctx, job.RunID); err != nil {
		slog.Error("failed to process run", "error", err, "run_id", job.RunID)
		errorMsg := err.Error()

		// Update run status to failed
		if err := w.updateRunStatus(ctx, job.RunID, runStatusFailed, &errorMsg); err != nil {
			slog.Error("failed to update run status to failed", "error", err, "run_id", job.RunID)
		}

		// Update job status
		requeue := job.Attempts < job.MaxAttempts
		if err := w.jobsRepo.MarkJobFailed(ctx, job.ID, errorMsg, requeue); err != nil {
			slog.Error("failed to mark job as failed", "error", err, "job_id", job.ID)
		}
		return err
	}

	// Success: update run status to completed
	if err := w.updateRunStatus(ctx, job.RunID, runStatusCompleted, nil); err != nil {
		slog.Error("failed to update run status to completed", "error", err, "run_id", job.RunID)
		w.jobsRepo.MarkJobFailed(ctx, job.ID, fmt.Sprintf("failed to update run status: %v", err), false)
		return err
	}

	// Mark job as done
	if err := w.jobsRepo.MarkJobDone(ctx, job.ID); err != nil {
		slog.Error("failed to mark job as done", "error", err, "job_id", job.ID)
		return err
	}

	slog.Info("job completed", "job_id", job.ID, "run_id", job.RunID, "worker_id", w.workerID)
	return nil
}

func (w *Worker) processRun(ctx context.Context, runID uuid.UUID) error {
	// Check if AI client is available
	if w.aiClient == nil {
		return fmt.Errorf("OPENAI_API_KEY missing")
	}

	// 1. Load the run
	runData, err := w.runsRepo.GetRunByID(ctx, runID)
	if err != nil {
		return fmt.Errorf("failed to load run: %w", err)
	}

	// 2. Load the resume
	resume, err := w.resumesRepo.GetResumeByID(ctx, runData.ResumeID)
	if err != nil {
		return fmt.Errorf("failed to load resume: %w", err)
	}

	resumeText := resume.ContentText
	jobText := runData.JobText

	// 3. Compute BM25 signals (stub for now)
	bm25Signals, err := bm25.Compute(resumeText, jobText)
	if err != nil {
		slog.Warn("BM25 computation failed, continuing without signals", "error", err, "run_id", runID)
		bm25Signals = nil
	}

	// 4. Generate ATS report and change plan via OpenAI
	atsReport, changePlan, err := w.aiClient.GenerateRunReport(ctx, resumeText, jobText, bm25Signals)
	if err != nil {
		return fmt.Errorf("failed to generate run report: %w", err)
	}

	// 5. Marshal to JSON
	atsReportJSON, err := json.Marshal(atsReport)
	if err != nil {
		return fmt.Errorf("failed to marshal ATS report: %w", err)
	}

	changePlanJSON, err := json.Marshal(changePlan)
	if err != nil {
		return fmt.Errorf("failed to marshal change plan: %w", err)
	}

	// 6. Persist into run_reports
	if w.reportsSvc != nil {
		if err := w.reportsSvc.UpsertRunReport(ctx, runID, atsReportJSON, changePlanJSON); err != nil {
			return fmt.Errorf("failed to upsert run report: %w", err)
		}
	}

	// Placeholder JSON for artifacts (LaTeX/PDF generation not implemented yet)
	resumeSpec := map[string]interface{}{
		"version":   "1.0",
		"sections":  []string{"placeholder section"},
		"timestamp": time.Now().Unix(),
	}

	resumeSpecJSON, err := json.Marshal(resumeSpec)
	if err != nil {
		return fmt.Errorf("failed to marshal resume spec: %w", err)
	}

	// Insert into run_artifacts
	const insertArtifactQ = `
INSERT INTO run_artifacts (run_id, resume_spec, latex_path, pdf_path)
VALUES ($1, $2, $3, $4)
ON CONFLICT (run_id) DO UPDATE
SET resume_spec = $2, latex_path = $3, pdf_path = $4, created_at = now()`

	latexPath := fmt.Sprintf("/generated/%s/resume.tex", runID.String())
	pdfPath := fmt.Sprintf("/generated/%s/resume.pdf", runID.String())

	_, err = w.db.Exec(ctx, insertArtifactQ, runID, resumeSpecJSON, latexPath, pdfPath)
	if err != nil {
		return fmt.Errorf("failed to insert run artifact: %w", err)
	}

	return nil
}

func (w *Worker) updateRunStatus(ctx context.Context, runID uuid.UUID, status string, errorMessage *string) error {
	const q = `
UPDATE runs
SET status = $2,
    error_message = $3,
    updated_at = now()
WHERE id = $1`

	_, err := w.db.Exec(ctx, q, runID, status, errorMessage)
	return err
}
