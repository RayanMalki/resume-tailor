package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"resume-tailor/internal/runreports"

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

type Worker struct {
	jobsRepo     *Repo
	db           *pgxpool.Pool
	workerID     string
	reportsSvc   *runreports.Service
}

func NewWorker(jobsRepo *Repo, db *pgxpool.Pool, workerID string, reportsSvc *runreports.Service) *Worker {
	return &Worker{
		jobsRepo:   jobsRepo,
		db:         db,
		workerID:   workerID,
		reportsSvc: reportsSvc,
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
	// MVP stub: pretend we generated something
	// In a real implementation, this would:
	// 1. Fetch the run and resume data
	// 2. Process the resume with the job text
	// 3. Generate ATS report and change plan
	// 4. Generate resume spec, LaTeX, and PDF
	// 5. Insert into run_reports and run_artifacts

	// For MVP, we'll insert placeholder JSON into run_reports and run_artifacts
	// This ensures the schema is consistent

	// Placeholder JSON for reports
	atsReport := map[string]interface{}{
		"score":  0.75,
		"notes":  []string{"placeholder"},
	}
	changePlan := map[string]interface{}{
		"changes": []string{"placeholder"},
	}

	atsReportJSON, err := json.Marshal(atsReport)
	if err != nil {
		return fmt.Errorf("failed to marshal ATS report: %w", err)
	}

	changePlanJSON, err := json.Marshal(changePlan)
	if err != nil {
		return fmt.Errorf("failed to marshal change plan: %w", err)
	}

	// Insert into run_reports using service
	if w.reportsSvc != nil {
		if err := w.reportsSvc.UpsertRunReport(ctx, runID, atsReportJSON, changePlanJSON); err != nil {
			return fmt.Errorf("failed to upsert run report: %w", err)
		}
	}

	// Placeholder JSON for artifacts
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
