# internal/jobs

- Relative path: `internal/jobs`
- Packages: `jobs`
- Source files: `queue.go`, `repo.go`, `types.go`, `worker.go`

## Functions

### `func (r *Repo) ClaimNextProcessRun(ctx context.Context, workerID string) (Job, error)`

- File: `repo.go`
- Purpose: Internal function supporting `internal/jobs` package workflows.

### `func (r *Repo) EnqueueProcessRun(ctx context.Context, runID uuid.UUID) (uuid.UUID, error)`

- File: `repo.go`
- Purpose: Internal function supporting `internal/jobs` package workflows.

### `func (r *Repo) MarkJobDone(ctx context.Context, jobID uuid.UUID) error`

- File: `repo.go`
- Purpose: Marks state transitions on an existing record.

### `func (r *Repo) MarkJobFailed(ctx context.Context, jobID uuid.UUID, errorMsg string, requeue bool) error`

- File: `repo.go`
- Purpose: Marks state transitions on an existing record.

### `func NewRepo(db *pgxpool.Pool) *Repo`

- File: `repo.go`
- Purpose: Creates and returns a new instance used by this package.

### `func NewWorker(jobsRepo *Repo, db *pgxpool.Pool, workerID string, reportsSvc *runreports.Service, runsRepo RunsRepo, resumesRepo *resumes.Repo, aiClient *ai.Client, artifactsSvc *artifacts.Service, pdfEnabled bool, tectonicBin string, jobTimeout time.Duration, disciplineMode string) *Worker`

- File: `worker.go`
- Purpose: Creates and returns a new instance used by this package.

### `func (w *Worker) Run(ctx context.Context) error`

- File: `worker.go`
- Purpose: Runs the package main workflow/loop.

### `func applyDisciplineSignals(signals *bm25.Signals, discipline profiles.Discipline, evidence []classifier.EvidenceTerm)`

- File: `worker.go`
- Purpose: Internal function supporting `internal/jobs` package workflows.

### `func clamp01(v float64) float64`

- File: `worker.go`
- Purpose: Internal helper for deterministic normalization/ordering.

### `func fallbackATSReport(addedKeywords, missingKeywords []string, resumeLang string) ai.ATSReport`

- File: `worker.go`
- Purpose: Fallback/safety helper used to keep processing resilient.

### `func (w *Worker) processNextJob(ctx context.Context) error`

- File: `worker.go`
- Purpose: Internal function supporting `internal/jobs` package workflows.

### `func (w *Worker) processRun(ctx context.Context, runID uuid.UUID) error`

- File: `worker.go`
- Purpose: Internal function supporting `internal/jobs` package workflows.

### `func safeErrorMessage(err error) string`

- File: `worker.go`
- Purpose: Fallback/safety helper used to keep processing resilient.

### `func toTermScores(evidence []classifier.EvidenceTerm) []bm25.TermScore`

- File: `worker.go`
- Purpose: Internal function supporting `internal/jobs` package workflows.

### `func (w *Worker) updateRunStatus(ctx context.Context, runID uuid.UUID, status string, errorMessage *string) error`

- File: `worker.go`
- Purpose: Internal function supporting `internal/jobs` package workflows.

