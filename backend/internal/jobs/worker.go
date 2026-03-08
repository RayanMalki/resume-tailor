package jobs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"resume-tailor/internal/ai"
	"resume-tailor/internal/artifacts"
	"resume-tailor/internal/crypto"
	"resume-tailor/internal/docx"
	"resume-tailor/internal/latex"
	"resume-tailor/internal/resumes"
	"resume-tailor/internal/runreports"
	"resume-tailor/internal/scoring/bm25"
	"resume-tailor/internal/scoring/classifier"
	"resume-tailor/internal/scoring/profiles"

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
	UpdateRunDiscipline(ctx context.Context, runID uuid.UUID, discipline string, confidence float64, source string) error
}

// UserKeyRepo provides per-user encrypted API key lookup.
type UserKeyRepo interface {
	GetEncryptedOpenAIKey(ctx context.Context, userID uuid.UUID) (*string, error)
}

// RunData represents the run data needed by the worker
type RunData struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	ResumeID         uuid.UUID
	JobText          string
	ProjectControls  []ai.ProjectControl
	Discipline       *string
	DisciplineScore  float64
	DisciplineSource string
	Status           string
	ErrorMessage     *string
}

type Worker struct {
	jobsRepo         *Repo
	db               *pgxpool.Pool
	workerID         string
	reportsSvc       *runreports.Service
	runsRepo         RunsRepo
	resumesRepo      *resumes.Repo
	aiClient         *ai.Client
	artifacts        *artifacts.Service
	pdfEnabled       bool
	tectonicBin      string
	jobTimeout       time.Duration
	disciplineMode   string
	userKeyRepo      UserKeyRepo
	apiKeyEncSecret  string
	openAIModel      string
}

func NewWorker(jobsRepo *Repo, db *pgxpool.Pool, workerID string, reportsSvc *runreports.Service, runsRepo RunsRepo, resumesRepo *resumes.Repo, aiClient *ai.Client, artifactsSvc *artifacts.Service, pdfEnabled bool, tectonicBin string, jobTimeout time.Duration, disciplineMode string, userKeyRepo UserKeyRepo, apiKeyEncSecret, openAIModel string) *Worker {
	if jobTimeout <= 0 {
		jobTimeout = defaultJobTimeout
	}
	switch strings.TrimSpace(strings.ToLower(disciplineMode)) {
	case "", "enforce":
		disciplineMode = "enforce"
	case "off", "observe":
		// valid
	default:
		disciplineMode = "enforce"
	}
	return &Worker{
		jobsRepo:        jobsRepo,
		db:              db,
		workerID:        workerID,
		reportsSvc:      reportsSvc,
		runsRepo:        runsRepo,
		resumesRepo:     resumesRepo,
		aiClient:        aiClient,
		artifacts:       artifactsSvc,
		pdfEnabled:      pdfEnabled,
		tectonicBin:     tectonicBin,
		jobTimeout:      jobTimeout,
		disciplineMode:  disciplineMode,
		userKeyRepo:     userKeyRepo,
		apiKeyEncSecret: apiKeyEncSecret,
		openAIModel:     openAIModel,
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
	// 1. Load the run
	runData, err := w.runsRepo.GetRunByID(ctx, runID)
	if err != nil {
		return fmt.Errorf("failed to load run: %w", err)
	}

	// Resolve effective AI client: prefer per-user key, fall back to global.
	effectiveClient := w.aiClient
	if w.userKeyRepo != nil && w.apiKeyEncSecret != "" && runData.UserID != uuid.Nil {
		if enc, err := w.userKeyRepo.GetEncryptedOpenAIKey(ctx, runData.UserID); err == nil && enc != nil {
			if plain, err := crypto.Decrypt(w.apiKeyEncSecret, *enc); err == nil && plain != "" {
				if c, err := ai.NewClientFromEnv(plain, w.openAIModel); err == nil {
					effectiveClient = c
					slog.Info("using per-user OpenAI key", "run_id", runID, "user_id", runData.UserID)
				}
			}
		}
	}

	if effectiveClient == nil {
		return fmt.Errorf("OPENAI_API_KEY missing")
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

	coverLetterPDFExists := false
	if w.pdfEnabled && w.artifacts != nil {
		if _, err := w.artifacts.GetByRunIDAndType(ctx, runID, artifacts.TypeCoverLetterPDF); err == nil {
			coverLetterPDFExists = true
		} else if err != nil && err != artifacts.ErrArtifactNotFound {
			return fmt.Errorf("failed to check existing cover letter pdf artifact: %w", err)
		}
	}

	coverLetterDOCXExists := false
	if w.artifacts != nil {
		if _, err := w.artifacts.GetByRunIDAndType(ctx, runID, artifacts.TypeCoverLetterDOCX); err == nil {
			coverLetterDOCXExists = true
		} else if err != nil && err != artifacts.ErrArtifactNotFound {
			return fmt.Errorf("failed to check existing cover letter docx artifact: %w", err)
		}
	}

	needsReasons := len(runData.ProjectControls) > 0
	if reportExists && latexExists && docxExists && coverLetterExists && coverLetterDOCXExists && (pdfExists || !w.pdfEnabled) && (coverLetterPDFExists || !w.pdfEnabled) && (!needsReasons || reasonsExists) {
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

	// 3. Resolve discipline and scoring profile.
	selectedDiscipline := profiles.DefaultDiscipline()
	selectedConfidence := 0.0
	selectedSource := profiles.SourceAuto
	lowConfidence := true
	var disciplineEvidence []classifier.EvidenceTerm

	if runData.Discipline != nil {
		if parsed, ok := profiles.ParseDiscipline(strings.TrimSpace(*runData.Discipline)); ok {
			selectedDiscipline = parsed
		}
	}
	if runData.DisciplineScore > 0 {
		selectedConfidence = clamp01(runData.DisciplineScore)
	}
	if src := strings.TrimSpace(strings.ToLower(runData.DisciplineSource)); src == string(profiles.SourceUserOverride) {
		selectedSource = profiles.SourceUserOverride
		selectedConfidence = 1
		lowConfidence = false
	} else {
		detected := classifier.Detect(resumeText, jobText)
		selectedDiscipline = detected.Discipline
		selectedConfidence = clamp01(detected.Confidence)
		selectedSource = detected.Source
		lowConfidence = detected.LowConfidence
		disciplineEvidence = detected.Evidence
	}

	// Persist selected discipline metadata for GET /runs/{runID}.
	if err := w.runsRepo.UpdateRunDiscipline(
		ctx,
		runID,
		string(selectedDiscipline),
		selectedConfidence,
		string(selectedSource),
	); err != nil {
		slog.Warn("failed to persist run discipline metadata", "run_id", runID, "error", err)
	}

	effectiveDiscipline := selectedDiscipline
	switch w.disciplineMode {
	case "off", "observe":
		effectiveDiscipline = profiles.DefaultDiscipline()
	}
	effectiveProfile := profiles.Get(effectiveDiscipline)

	slog.Info(
		"discipline_classification",
		"run_id", runID,
		"mode", w.disciplineMode,
		"selected", selectedDiscipline,
		"effective", effectiveDiscipline,
		"confidence", selectedConfidence,
		"low_confidence", lowConfidence,
		"source", selectedSource,
	)

	// 4. Compute BM25 signals on ORIGINAL resume
	var bm25Signals *bm25.Signals
	signals, err := bm25.ComputeWithProfile(resumeText, jobText, effectiveProfile)
	if err != nil {
		slog.Warn("BM25 computation failed, continuing without signals", "error", err, "run_id", runID)
	} else {
		applyDisciplineSignals(&signals, selectedDiscipline, disciplineEvidence)
		bm25Signals = &signals
	}

	// 5. Generate resume spec + LaTeX FIRST (before the report)
	var latexDoc string
	var spec ai.ResumeSpec
	var specGenerated bool
	if w.artifacts != nil && (!latexExists || !docxExists) {
		spec, err = effectiveClient.GenerateResumeSpec(ctx, resumeText, jobText, bm25Signals, runData.ProjectControls, ai.DisciplineContext{
			Discipline:    string(effectiveDiscipline),
			RoleFocus:     effectiveProfile.PromptHints.RoleFocus,
			EvidenceFocus: effectiveProfile.PromptHints.EvidenceFocus,
			ActionVerbs:   effectiveProfile.PromptHints.ActionVerbs,
		})
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
			spec, err = effectiveClient.GenerateResumeSpec(ctx, resumeText, jobText, bm25Signals, runData.ProjectControls, ai.DisciplineContext{
				Discipline:    string(effectiveDiscipline),
				RoleFocus:     effectiveProfile.PromptHints.RoleFocus,
				EvidenceFocus: effectiveProfile.PromptHints.EvidenceFocus,
				ActionVerbs:   effectiveProfile.PromptHints.ActionVerbs,
			})
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
			ts, err := bm25.ComputeWithProfile(tailoredText, jobText, effectiveProfile)
			if err != nil {
				slog.Warn("BM25 computation on tailored resume failed", "error", err, "run_id", runID)
				if bm25Signals != nil {
					tailoredSignals = *bm25Signals
				}
			} else {
				applyDisciplineSignals(&ts, selectedDiscipline, disciplineEvidence)
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
		atsReport, _, err := effectiveClient.GenerateRunReport(ctx, addedKeywords, missingNames, resumeLang, ai.DisciplineContext{
			Discipline: string(effectiveDiscipline),
			RoleFocus:  effectiveProfile.PromptHints.RoleFocus,
		})
		if err != nil {
			// Report generation should not fail the entire run; fallback to deterministic notes.
			slog.Warn("failed to generate run report; using deterministic fallback", "run_id", runID, "error", err)
			atsReport = fallbackATSReport(addedKeywords, missingNames, resumeLang)
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
		originalMissingCount := 0
		if bm25Signals != nil {
			originalMissingCount = len(bm25Signals.MissingJobTerms)
		}
		slog.Info(
			"discipline_run_metrics",
			"run_id", runID,
			"discipline", selectedDiscipline,
			"source", selectedSource,
			"low_confidence", lowConfidence,
			"missing_before", originalMissingCount,
			"missing_after", len(tailoredSignals.MissingJobTerms),
			"mode", w.disciplineMode,
		)

		payload := reportPayload{
			ReportVersion:        2,
			GeneratedAt:          time.Now().UTC().Format(time.RFC3339),
			BM25Signals:          tailoredSignals,
			ATSReport:            atsReport,
			ChangePlan:           changePlan,
			Discipline:           string(selectedDiscipline),
			DisciplineConfidence: selectedConfidence,
			DisciplineSource:     string(selectedSource),
			LowConfidence:        lowConfidence,
			DisciplineEvidence:   toTermScores(disciplineEvidence),
			CategoryCoverage:     tailoredSignals.CategoryCoverage,
			ProfileVersion:       profiles.Version(),
			ScoringDiscipline:    string(effectiveDiscipline),
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

	// 6-8. Non-critical artifacts are generated concurrently to reduce wall time.
	if w.artifacts != nil {
		var wg sync.WaitGroup

		// Compile PDF
		if w.pdfEnabled && !pdfExists && latexDoc != "" {
			wg.Add(1)
			go func() {
				defer wg.Done()
				pdfBytes, err := latex.CompilePDF(ctx, w.tectonicBin, latexDoc)
				if err != nil {
					slog.Warn("failed to compile resume pdf; continuing without pdf artifact", "run_id", runID, "error", err)
					return
				}
				encoded := base64.StdEncoding.EncodeToString(pdfBytes)
				if err := w.artifacts.InsertIfNotExists(ctx, runID, artifacts.TypeResumePDF, encoded); err != nil {
					slog.Warn("failed to store resume pdf artifact; continuing", "run_id", runID, "error", err)
				}
			}()
		}

		// Generate project reasons (if project controls were used)
		if needsReasons && !reasonsExists && latexDoc != "" {
			wg.Add(1)
			go func() {
				defer wg.Done()
				reasons, err := effectiveClient.GenerateProjectReasons(ctx, resumeText, jobText, latexDoc, bm25Signals, runData.ProjectControls)
				if err != nil {
					slog.Warn("failed to generate project reasons; continuing without project reasons artifact", "run_id", runID, "error", err)
					return
				}
				reasonsJSON, err := json.Marshal(reasons)
				if err != nil {
					slog.Warn("failed to marshal project reasons; continuing without project reasons artifact", "run_id", runID, "error", err)
					return
				}
				if err := w.artifacts.InsertIfNotExists(ctx, runID, artifacts.TypeProjectReasons, string(reasonsJSON)); err != nil {
					slog.Warn("failed to store project reasons artifact; continuing", "run_id", runID, "error", err)
				}
			}()
		}

		// Generate cover letter (best effort)
		if !coverLetterExists || !coverLetterDOCXExists || (w.pdfEnabled && !coverLetterPDFExists) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				resumeLang := "French"
				if specGenerated && strings.TrimSpace(spec.Language) != "" {
					resumeLang = spec.Language
				}

				var coverLetter string
				if !coverLetterExists {
					cl, err := effectiveClient.GenerateCoverLetter(ctx, resumeText, jobText, bm25Signals, resumeLang, ai.DisciplineContext{
						Discipline:    string(effectiveDiscipline),
						RoleFocus:     effectiveProfile.PromptHints.RoleFocus,
						EvidenceFocus: effectiveProfile.PromptHints.EvidenceFocus,
					})
					if err != nil {
						slog.Warn("failed to generate cover letter; continuing without cover letter artifact", "run_id", runID, "error", err)
						return
					}
					coverLetter = cl
					if err := w.artifacts.InsertIfNotExists(ctx, runID, artifacts.TypeCoverLetter, coverLetter); err != nil {
						slog.Warn("failed to store cover letter artifact; continuing", "run_id", runID, "error", err)
					}
				} else {
					existing, err := w.artifacts.GetByRunIDAndType(ctx, runID, artifacts.TypeCoverLetter)
					if err != nil {
						slog.Warn("failed to load existing cover letter; skipping pdf/docx generation", "run_id", runID, "error", err)
						return
					}
					coverLetter = existing.Content
				}

				// Build contact string from spec
				clContact := buildContactString(spec.Contact)
				clName := spec.Name

				// Generate cover letter DOCX
				if !coverLetterDOCXExists {
					clDocxBytes, err := docx.RenderCoverLetter(clName, clContact, coverLetter)
					if err != nil {
						slog.Warn("failed to render cover letter docx; continuing", "run_id", runID, "error", err)
					} else {
						clDocxB64 := base64.StdEncoding.EncodeToString(clDocxBytes)
						if err := w.artifacts.InsertIfNotExists(ctx, runID, artifacts.TypeCoverLetterDOCX, clDocxB64); err != nil {
							slog.Warn("failed to store cover letter docx artifact; continuing", "run_id", runID, "error", err)
						}
					}
				}

				// Generate cover letter PDF
				if w.pdfEnabled && !coverLetterPDFExists {
					clLatex := latex.RenderCoverLetter(clName, clContact, coverLetter)
					pdfBytes, err := latex.CompilePDF(ctx, w.tectonicBin, clLatex)
					if err != nil {
						slog.Warn("failed to compile cover letter pdf; continuing", "run_id", runID, "error", err)
					} else {
						clPdfB64 := base64.StdEncoding.EncodeToString(pdfBytes)
						if err := w.artifacts.InsertIfNotExists(ctx, runID, artifacts.TypeCoverLetterPDF, clPdfB64); err != nil {
							slog.Warn("failed to store cover letter pdf artifact; continuing", "run_id", runID, "error", err)
						}
					}
				}
			}()
		}

		wg.Wait()
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
	ReportVersion        int                `json:"report_version"`
	GeneratedAt          string             `json:"generated_at"`
	BM25Signals          bm25.Signals       `json:"bm25_signals"`
	ATSReport            ai.ATSReport       `json:"ats_report"`
	ChangePlan           ai.ChangePlan      `json:"change_plan"`
	Discipline           string             `json:"discipline,omitempty"`
	DisciplineConfidence float64            `json:"discipline_confidence,omitempty"`
	DisciplineSource     string             `json:"discipline_source,omitempty"`
	LowConfidence        bool               `json:"low_confidence,omitempty"`
	DisciplineEvidence   []bm25.TermScore   `json:"discipline_evidence,omitempty"`
	CategoryCoverage     map[string]float64 `json:"category_coverage,omitempty"`
	ProfileVersion       string             `json:"profile_version,omitempty"`
	ScoringDiscipline    string             `json:"scoring_discipline,omitempty"`
}

func applyDisciplineSignals(signals *bm25.Signals, discipline profiles.Discipline, evidence []classifier.EvidenceTerm) {
	if signals == nil {
		return
	}
	signals.Discipline = string(discipline)
	signals.ProfileVersion = profiles.Version()
	signals.DisciplineEvidence = toTermScores(evidence)
}

func toTermScores(evidence []classifier.EvidenceTerm) []bm25.TermScore {
	if len(evidence) == 0 {
		return nil
	}
	out := make([]bm25.TermScore, 0, len(evidence))
	for _, item := range evidence {
		if strings.TrimSpace(item.Term) == "" {
			continue
		}
		out = append(out, bm25.TermScore{
			Term:  item.Term,
			Score: item.Score,
		})
	}
	return out
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func fallbackATSReport(addedKeywords, missingKeywords []string, resumeLang string) ai.ATSReport {
	isFrench := strings.Contains(strings.ToLower(resumeLang), "fr")
	var summary string
	notes := make([]string, 0, 3)

	if isFrench {
		if len(addedKeywords) > 0 {
			summary = fmt.Sprintf("Le CV a ete adapte avec %d mots-cles pertinents. Certains mots-cles restent absents et peuvent etre ajoutes si l'experience est verifiable.", len(addedKeywords))
		} else {
			summary = "Le CV etait deja relativement aligne. Quelques mots-cles importants restent absents et peuvent reduire la compatibilite ATS."
		}
		notes = append(notes, fmt.Sprintf("Mots-cles ajoutes: %d", len(addedKeywords)))
		notes = append(notes, fmt.Sprintf("Mots-cles manquants: %d", len(missingKeywords)))
		notes = append(notes, "Rapport genere en mode secours a cause d'une reponse IA invalide.")
	} else {
		if len(addedKeywords) > 0 {
			summary = fmt.Sprintf("The resume was tailored with %d relevant keywords. Some keywords are still missing and can be added when backed by real experience.", len(addedKeywords))
		} else {
			summary = "The resume was already fairly aligned. Some high-signal keywords are still missing and may reduce ATS match quality."
		}
		notes = append(notes, fmt.Sprintf("Added keywords: %d", len(addedKeywords)))
		notes = append(notes, fmt.Sprintf("Missing keywords: %d", len(missingKeywords)))
		notes = append(notes, "Report generated in fallback mode due to invalid AI JSON output.")
	}

	return ai.ATSReport{
		Notes:   notes,
		Summary: summary,
	}
}

func buildContactString(contact []string) string {
	var parts []string
	for _, c := range contact {
		if trimmed := strings.TrimSpace(c); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, " · ")
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
