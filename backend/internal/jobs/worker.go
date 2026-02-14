package jobs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"resume-tailor/internal/ai"
	"resume-tailor/internal/artifacts"
	"resume-tailor/internal/docx"
	"resume-tailor/internal/latex"
	"resume-tailor/internal/resumes"
	"resume-tailor/internal/runreports"
	"resume-tailor/internal/scoring/bm25"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const pollInterval = 1 * time.Second

const (
	defaultJobTimeout = 15 * time.Minute
	finalizeTimeout   = 15 * time.Second
)

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
	jobTimeout  time.Duration
}

func NewWorker(jobsRepo *Repo, db *pgxpool.Pool, workerID string, reportsSvc *runreports.Service, runsRepo RunsRepo, resumesRepo *resumes.Repo, aiClient *ai.Client, artifactsSvc *artifacts.Service, pdfEnabled bool, tectonicBin string, jobTimeout time.Duration) *Worker {
	if jobTimeout <= 0 {
		jobTimeout = defaultJobTimeout
	}
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
		jobTimeout:  jobTimeout,
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
	jobCtx, cancel := context.WithTimeout(ctx, w.jobTimeout)
	defer cancel()

	// Update run status to running
	if err := w.updateRunStatus(jobCtx, job.RunID, runStatusRunning, nil); err != nil {
		slog.Error("failed to update run status to running", "error", err, "run_id", job.RunID)
		// Mark job as failed
		finalizeCtx, finalizeCancel := context.WithTimeout(context.Background(), finalizeTimeout)
		defer finalizeCancel()
		w.jobsRepo.MarkJobFailed(finalizeCtx, job.ID, fmt.Sprintf("failed to update run status: %v", err), job.Attempts < job.MaxAttempts)
		return err
	}

	// Process the run
	if err := w.processRun(jobCtx, job.RunID); err != nil {
		slog.Error("failed to process run", "error", err, "run_id", job.RunID)
		errorMsg := safeErrorMessage(err)
		finalizeCtx, finalizeCancel := context.WithTimeout(context.Background(), finalizeTimeout)
		defer finalizeCancel()

		// Update run status to failed
		if err := w.updateRunStatus(finalizeCtx, job.RunID, runStatusFailed, &errorMsg); err != nil {
			slog.Error("failed to update run status to failed", "error", err, "run_id", job.RunID)
		}

		// Update job status
		requeue := job.Attempts < job.MaxAttempts
		if err := w.jobsRepo.MarkJobFailed(finalizeCtx, job.ID, errorMsg, requeue); err != nil {
			slog.Error("failed to mark job as failed", "error", err, "job_id", job.ID)
		}
		return err
	}

	finalizeCtx, finalizeCancel := context.WithTimeout(context.Background(), finalizeTimeout)
	defer finalizeCancel()

	// Success: update run status to succeeded
	if err := w.updateRunStatus(finalizeCtx, job.RunID, runStatusSucceeded, nil); err != nil {
		slog.Error("failed to update run status to succeeded", "error", err, "run_id", job.RunID)
		w.jobsRepo.MarkJobFailed(finalizeCtx, job.ID, fmt.Sprintf("failed to update run status: %v", err), false)
		return err
	}

	// Mark job as done
	if err := w.jobsRepo.MarkJobDone(finalizeCtx, job.ID); err != nil {
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

	docxExists := false
	if w.artifacts != nil {
		if _, err := w.artifacts.GetByRunIDAndType(ctx, runID, artifacts.TypeResumeDOCX); err == nil {
			docxExists = true
		} else if err != nil && err != artifacts.ErrArtifactNotFound {
			return fmt.Errorf("failed to check existing docx artifact: %w", err)
		}
	}

	coverLetterExists := false
	if w.artifacts != nil {
		if _, err := w.artifacts.GetByRunIDAndType(ctx, runID, artifacts.TypeCoverLetter); err == nil {
			coverLetterExists = true
		} else if err != nil && err != artifacts.ErrArtifactNotFound {
			return fmt.Errorf("failed to check existing cover letter artifact: %w", err)
		}
	}

	needsReasons := len(runData.ProjectControls) > 0
	if reportExists && latexExists && docxExists && coverLetterExists && (pdfExists || !w.pdfEnabled) && (!needsReasons || reasonsExists) {
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

	// 3. Compute BM25 signals on ORIGINAL resume
	var bm25Signals *bm25.Signals
	signals, err := bm25.Compute(resumeText, jobText)
	if err != nil {
		slog.Warn("BM25 computation failed, continuing without signals", "error", err, "run_id", runID)
	} else {
		bm25Signals = &signals
	}

	// 4. Generate resume spec + LaTeX FIRST (before the report)
	var latexDoc string
	var spec ai.ResumeSpec
	var specGenerated bool
	if w.artifacts != nil && (!latexExists || !docxExists) {
		if w.aiClient == nil {
			return fmt.Errorf("ai client is not configured")
		}
		spec, err = w.aiClient.GenerateResumeSpec(ctx, resumeText, jobText, bm25Signals, runData.ProjectControls)
		if err != nil {
			return fmt.Errorf("failed to generate resume spec: %w", err)
		}
		specGenerated = true
		if !latexExists {
			latexDoc = latex.RenderResume(spec)
			if err := w.artifacts.InsertIfNotExists(ctx, runID, artifacts.TypeResumeLatex, latexDoc); err != nil {
				return fmt.Errorf("failed to store resume latex: %w", err)
			}
			latexExists = true
		}
	} else if latexExists && w.artifacts != nil {
		existing, err := w.artifacts.GetByRunIDAndType(ctx, runID, artifacts.TypeResumeLatex)
		if err != nil {
			return fmt.Errorf("failed to load resume latex: %w", err)
		}
		latexDoc = existing.Content
	}

	// 4.1 Generate DOCX artifact for editable resume output.
	if w.artifacts != nil && !docxExists {
		if !specGenerated {
			spec, err = w.aiClient.GenerateResumeSpec(ctx, resumeText, jobText, bm25Signals, runData.ProjectControls)
			if err != nil {
				return fmt.Errorf("failed to generate resume spec for docx: %w", err)
			}
			specGenerated = true
		}
		docxBytes, err := docx.RenderResume(spec)
		if err != nil {
			return fmt.Errorf("failed to render resume docx: %w", err)
		}
		encoded := base64.StdEncoding.EncodeToString(docxBytes)
		if err := w.artifacts.InsertIfNotExists(ctx, runID, artifacts.TypeResumeDOCX, encoded); err != nil {
			return fmt.Errorf("failed to store resume docx: %w", err)
		}
	}

	// 5. Compute BM25 on TAILORED resume and generate report
	if !reportExists {
		// Compute BM25 on the tailored resume text for an accurate score
		var tailoredSignals bm25.Signals
		if specGenerated {
			tailoredText := ai.ResumeSpecToText(spec)
			ts, err := bm25.Compute(tailoredText, jobText)
			if err != nil {
				slog.Warn("BM25 computation on tailored resume failed", "error", err, "run_id", runID)
				if bm25Signals != nil {
					tailoredSignals = *bm25Signals
				}
			} else {
				tailoredSignals = ts
			}
		} else if bm25Signals != nil {
			tailoredSignals = *bm25Signals
		}

		// Compute ATS score from IDF-weighted keyword coverage.
		atsScore := tailoredSignals.Score

		// Build change plan PROGRAMMATICALLY by comparing original vs tailored BM25.
		// This is 100% accurate — no AI hallucination possible.
		originalOverlap := make(map[string]bool)
		if bm25Signals != nil {
			for _, t := range bm25Signals.OverlapTerms {
				originalOverlap[t] = true
			}
		}
		// Keywords that moved from missing → matched = real additions
		var addedKeywords []string
		for _, t := range tailoredSignals.OverlapTerms {
			if !originalOverlap[t] {
				addedKeywords = append(addedKeywords, t)
			}
		}
		// Keywords still missing
		missingNames := make([]string, 0, len(tailoredSignals.MissingJobTerms))
		for _, t := range tailoredSignals.MissingJobTerms {
			missingNames = append(missingNames, t.Term)
		}

		// Detect resume language
		resumeLang := "French"
		if specGenerated && spec.Language != "" {
			resumeLang = spec.Language
		}

		// Generate summary + notes from AI (no change plan — that's computed above)
		atsReport, _, err := w.aiClient.GenerateRunReport(ctx, addedKeywords, missingNames, resumeLang)
		if err != nil {
			return fmt.Errorf("failed to generate run report: %w", err)
		}

		// Set score from IDF-weighted BM25 keyword coverage.
		atsReport.Score = atsScore

		// Build change plan programmatically from BM25 diff (100% accurate, no AI lies)
		var changes []string
		if len(addedKeywords) > 0 {
			changes = append(changes, fmt.Sprintf("Mots-clés ajoutés au CV : %s", strings.Join(addedKeywords, ", ")))
		}
		if len(missingNames) > 0 && len(missingNames) <= 10 {
			changes = append(changes, fmt.Sprintf("Mots-clés toujours manquants : %s", strings.Join(missingNames, ", ")))
		} else if len(missingNames) > 10 {
			changes = append(changes, fmt.Sprintf("Mots-clés toujours manquants : %s, et %d autres", strings.Join(missingNames[:10], ", "), len(missingNames)-10))
		}
		if len(addedKeywords) == 0 {
			changes = append(changes, "Le CV original était déjà bien adapté au poste — aucun mot-clé significatif n'a été ajouté.")
		}
		changePlan := ai.ChangePlan{Changes: changes}

		payload := reportPayload{
			ReportVersion: 1,
			GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
			BM25Signals:   tailoredSignals,
			ATSReport:     atsReport,
			ChangePlan:    changePlan,
		}

		atsReportJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal ATS report: %w", err)
		}

		changePlanJSON, err := json.Marshal(changePlan)
		if err != nil {
			return fmt.Errorf("failed to marshal change plan: %w", err)
		}

		if w.reportsSvc != nil {
			if err := w.reportsSvc.UpsertRunReport(ctx, runID, atsReportJSON, changePlanJSON); err != nil {
				return fmt.Errorf("failed to upsert run report: %w", err)
			}
		}
	}

	// 6. Compile PDF
	if w.pdfEnabled && !pdfExists && w.artifacts != nil && latexDoc != "" {
		pdfBytes, err := latex.CompilePDF(ctx, w.tectonicBin, latexDoc)
		if err != nil {
			slog.Warn("failed to compile resume pdf; continuing without pdf artifact", "run_id", runID, "error", err)
		} else {
			encoded := base64.StdEncoding.EncodeToString(pdfBytes)
			if err := w.artifacts.InsertIfNotExists(ctx, runID, artifacts.TypeResumePDF, encoded); err != nil {
				return fmt.Errorf("failed to store resume pdf: %w", err)
			}
		}
	}

	// 7. Generate project reasons (if project controls were used)
	if needsReasons && !reasonsExists && w.artifacts != nil && latexDoc != "" {
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

	// 8. Generate cover letter (best effort; does not fail the run)
	if w.artifacts != nil && !coverLetterExists {
		resumeLang := "French"
		if specGenerated && strings.TrimSpace(spec.Language) != "" {
			resumeLang = spec.Language
		}
		coverLetter, err := w.aiClient.GenerateCoverLetter(ctx, resumeText, jobText, bm25Signals, resumeLang)
		if err != nil {
			slog.Warn("failed to generate cover letter; continuing without cover letter artifact", "run_id", runID, "error", err)
		} else if err := w.artifacts.InsertIfNotExists(ctx, runID, artifacts.TypeCoverLetter, coverLetter); err != nil {
			slog.Warn("failed to store cover letter artifact; continuing", "run_id", runID, "error", err)
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
	if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "context deadline exceeded") {
		return "processing_timed_out"
	}
	if errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "context canceled") {
		return "processing_canceled"
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
