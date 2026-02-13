package jobs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"resume-tailor/internal/ai"
	"resume-tailor/internal/artifacts"
	"resume-tailor/internal/latex"
	"resume-tailor/internal/resumes"
	"resume-tailor/internal/runreports"
	"resume-tailor/internal/scoring/bm25"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const pollInterval = 1 * time.Second
const jobTimeout = 5 * time.Minute

const (
	runStatusQueued    = "queued"
	runStatusRunning   = "running"
	runStatusFailed    = "failed"
	runStatusSucceeded = "succeeded"
)

// RunsRepo is an interface to avoid import cycle with runs package
type RunsRepo interface {
	GetRunByID(ctx context.Context, runID uuid.UUID) (RunData, error)
}

// RunData represents the run data needed by the worker
type RunData struct {
	ID              uuid.UUID
	ResumeID        uuid.UUID
	JobText         string
	ProjectControls []ai.ProjectControl
	Status          string
	ErrorMessage    *string
}

type Worker struct {
	jobsRepo    *Repo
	db          *pgxpool.Pool
	workerID    string
	reportsSvc  *runreports.Service
	runsRepo    RunsRepo
	resumesRepo *resumes.Repo
	aiClient    *ai.Client
	artifacts   *artifacts.Service
	pdfEnabled  bool
	tectonicBin string
}

func NewWorker(jobsRepo *Repo, db *pgxpool.Pool, workerID string, reportsSvc *runreports.Service, runsRepo RunsRepo, resumesRepo *resumes.Repo, aiClient *ai.Client, artifactsSvc *artifacts.Service, pdfEnabled bool, tectonicBin string) *Worker {
	return &Worker{
		jobsRepo:    jobsRepo,
		db:          db,
		workerID:    workerID,
		reportsSvc:  reportsSvc,
		runsRepo:    runsRepo,
		resumesRepo: resumesRepo,
		aiClient:    aiClient,
		artifacts:   artifactsSvc,
		pdfEnabled:  pdfEnabled,
		tectonicBin: tectonicBin,
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

	// Enforce a timeout on the entire job so a hung LLM call or PDF compile
	// cannot block the worker forever.
	jobCtx, cancel := context.WithTimeout(ctx, jobTimeout)
	defer cancel()

	// Update run status to running
	if err := w.updateRunStatus(jobCtx, job.RunID, runStatusRunning, nil); err != nil {
		slog.Error("failed to update run status to running", "error", err, "run_id", job.RunID)
		// Mark job as failed
		w.jobsRepo.MarkJobFailed(jobCtx, job.ID, fmt.Sprintf("failed to update run status: %v", err), job.Attempts < job.MaxAttempts)
		return err
	}

	// Process the run
	if err := w.processRun(jobCtx, job.RunID); err != nil {
		slog.Error("failed to process run", "error", err, "run_id", job.RunID)
		errorMsg := safeErrorMessage(err)

		// Update run status to failed
		if err := w.updateRunStatus(jobCtx, job.RunID, runStatusFailed, &errorMsg); err != nil {
			slog.Error("failed to update run status to failed", "error", err, "run_id", job.RunID)
		}

		// Update job status
		requeue := job.Attempts < job.MaxAttempts
		if err := w.jobsRepo.MarkJobFailed(jobCtx, job.ID, errorMsg, requeue); err != nil {
			slog.Error("failed to mark job as failed", "error", err, "job_id", job.ID)
		}
		return err
	}

	// Success: update run status to succeeded
	if err := w.updateRunStatus(jobCtx, job.RunID, runStatusSucceeded, nil); err != nil {
		slog.Error("failed to update run status to succeeded", "error", err, "run_id", job.RunID)
		w.jobsRepo.MarkJobFailed(jobCtx, job.ID, fmt.Sprintf("failed to update run status: %v", err), false)
		return err
	}

	// Mark job as done
	if err := w.jobsRepo.MarkJobDone(jobCtx, job.ID); err != nil {
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

	reportExists := false
	if w.reportsSvc != nil {
		if _, err := w.reportsSvc.GetRunReportByRunID(ctx, runID); err == nil {
			reportExists = true
		} else if err != nil && err != runreports.ErrRunReportNotFound {
			return fmt.Errorf("failed to check existing run report: %w", err)
		}
	}

	latexExists := false
	if w.artifacts != nil {
		if _, err := w.artifacts.GetByRunIDAndType(ctx, runID, artifacts.TypeResumeLatex); err == nil {
			latexExists = true
		} else if err != nil && err != artifacts.ErrArtifactNotFound {
			return fmt.Errorf("failed to check existing latex artifact: %w", err)
		}
	}

	pdfExists := false
	if w.pdfEnabled && w.artifacts != nil {
		if _, err := w.artifacts.GetByRunIDAndType(ctx, runID, artifacts.TypeResumePDF); err == nil {
			pdfExists = true
		} else if err != nil && err != artifacts.ErrArtifactNotFound {
			return fmt.Errorf("failed to check existing pdf artifact: %w", err)
		}
	}

	reasonsExists := false
	if w.artifacts != nil {
		if _, err := w.artifacts.GetByRunIDAndType(ctx, runID, artifacts.TypeProjectReasons); err == nil {
			reasonsExists = true
		} else if err != nil && err != artifacts.ErrArtifactNotFound {
			return fmt.Errorf("failed to check existing project reasons artifact: %w", err)
		}
	}

	needsReasons := len(runData.ProjectControls) > 0
	if reportExists && latexExists && (pdfExists || !w.pdfEnabled) && (!needsReasons || reasonsExists) {
		slog.Info("run report and artifacts already exist, skipping", "run_id", runID)
		return nil
	}

	// 2. Load the resume
	resume, err := w.resumesRepo.GetResumeByID(ctx, runData.ResumeID)
	if err != nil {
		return fmt.Errorf("failed to load resume: %w", err)
	}

	resumeText := resume.ContentText
	jobText := runData.JobText

	// 3. Compute BM25 signals (stub for now)
	var bm25Signals *bm25.Signals
	signals, err := bm25.Compute(resumeText, jobText)
	if err != nil {
		slog.Warn("BM25 computation failed, continuing without signals", "error", err, "run_id", runID)
	} else {
		bm25Signals = &signals
	}

	// 4. Generate ATS report and change plan via OpenAI
	var atsReport ai.ATSReport
	var changePlan ai.ChangePlan
	if !reportExists {
		atsReport, changePlan, err = w.aiClient.GenerateRunReport(ctx, resumeText, jobText, bm25Signals)
		if err != nil {
			return fmt.Errorf("failed to generate run report: %w", err)
		}
	}

	reportSignals := bm25.Signals{}
	if bm25Signals != nil {
		reportSignals = *bm25Signals
	}

	if !reportExists {
		reportPayload := reportPayload{
			ReportVersion: 1,
			GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
			BM25Signals:   reportSignals,
			ATSReport:     atsReport,
			ChangePlan:    changePlan,
		}

		// 5. Marshal to JSON
		atsReportJSON, err := json.Marshal(reportPayload)
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
	}

	if w.artifacts != nil && (!latexExists || (w.pdfEnabled && !pdfExists)) {
		if w.aiClient == nil {
			return fmt.Errorf("ai client is not configured")
		}

		var latexDoc string
		if latexExists {
			existing, err := w.artifacts.GetByRunIDAndType(ctx, runID, artifacts.TypeResumeLatex)
			if err != nil {
				return fmt.Errorf("failed to load resume latex: %w", err)
			}
			latexDoc = existing.Content
		} else {
			spec, err := w.aiClient.GenerateResumeSpec(ctx, resumeText, jobText, bm25Signals, runData.ProjectControls)
			if err != nil {
				return fmt.Errorf("failed to generate resume spec: %w", err)
			}
			latexDoc = latex.RenderResume(spec)
			if err := w.artifacts.InsertIfNotExists(ctx, runID, artifacts.TypeResumeLatex, latexDoc); err != nil {
				return fmt.Errorf("failed to store resume latex: %w", err)
			}
		}

		if w.pdfEnabled && !pdfExists {
			pdfBytes, err := latex.CompilePDF(ctx, w.tectonicBin, latexDoc)
			if err != nil {
				// PDF generation is best-effort; keep run successful if LaTeX is ready.
				slog.Warn("failed to compile resume pdf; continuing without pdf artifact", "run_id", runID, "error", err)
			} else {
				encoded := base64.StdEncoding.EncodeToString(pdfBytes)
				if err := w.artifacts.InsertIfNotExists(ctx, runID, artifacts.TypeResumePDF, encoded); err != nil {
					return fmt.Errorf("failed to store resume pdf: %w", err)
				}
			}
		}

		if needsReasons && !reasonsExists {
			reasons, err := w.aiClient.GenerateProjectReasons(ctx, resumeText, jobText, latexDoc, bm25Signals, runData.ProjectControls)
			if err != nil {
				slog.Warn("failed to generate project reasons; continuing without project reasons artifact", "run_id", runID, "error", err)
			} else {
				reasonsJSON, err := json.Marshal(reasons)
				if err != nil {
					slog.Warn("failed to marshal project reasons; continuing without project reasons artifact", "run_id", runID, "error", err)
				} else if err := w.artifacts.InsertIfNotExists(ctx, runID, artifacts.TypeProjectReasons, string(reasonsJSON)); err != nil {
					slog.Warn("failed to store project reasons artifact; continuing", "run_id", runID, "error", err)
				}
			}
		}
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

type reportPayload struct {
	ReportVersion int           `json:"report_version"`
	GeneratedAt   string        `json:"generated_at"`
	BM25Signals   bm25.Signals  `json:"bm25_signals"`
	ATSReport     ai.ATSReport  `json:"ats_report"`
	ChangePlan    ai.ChangePlan `json:"change_plan"`
}

func safeErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "OPENAI_API_KEY"):
		return "ai_client_unavailable"
	case strings.Contains(msg, "generate run report"):
		return "report_generation_failed"
	case strings.Contains(msg, "generate resume latex"):
		return "latex_generation_failed"
	case strings.Contains(msg, "generate resume spec"):
		return "resume_spec_generation_failed"
	case strings.Contains(msg, "project reasons"):
		return "project_reasons_generation_failed"
	case strings.Contains(msg, "compile resume pdf"):
		return "pdf_generation_failed"
	case strings.Contains(msg, "load resume"):
		return "resume_load_failed"
	case strings.Contains(msg, "load run"):
		return "run_load_failed"
	default:
		return "processing_failed"
	}
}
