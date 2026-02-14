# internal/runs

- Relative path: `internal/runs`
- Packages: `runs`
- Source files: `repo.go`, `service.go`, `types.go`

## Functions

### `func (r *Repo) CreateRun(ctx context.Context, userID, resumeID uuid.UUID, jobText string, projectControls []ProjectControl, discipline *Discipline, disciplineScore float64, disciplineSource DisciplineSource) (Run, error)`

- File: `repo.go`
- Purpose: Creates a new record or resource and returns the result.

### `func (r *Repo) GetRunByID(ctx context.Context, runID uuid.UUID) (Run, error)`

- File: `repo.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func (r *Repo) ListRunsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Run, error)`

- File: `repo.go`
- Purpose: Fetches a paginated or ordered list of resources.

### `func NewRepo(db *pgxpool.Pool) *Repo`

- File: `repo.go`
- Purpose: Creates and returns a new instance used by this package.

### `func (r *Repo) UpdateRunDiscipline(ctx context.Context, runID uuid.UUID, discipline Discipline, confidence float64, source DisciplineSource) error`

- File: `repo.go`
- Purpose: Updates persisted state for an existing resource.

### `func (r *Repo) UpdateRunStatus(ctx context.Context, runID uuid.UUID, status string, errorMessage *string) error`

- File: `repo.go`
- Purpose: Updates persisted state for an existing resource.

### `func marshalProjectControls(controls []ProjectControl) ([]byte, error)`

- File: `repo.go`
- Purpose: Internal function supporting `internal/runs` package workflows.

### `func unmarshalProjectControls(raw []byte) []ProjectControl`

- File: `repo.go`
- Purpose: Internal function supporting `internal/runs` package workflows.

### `func (s *Service) CreateRun(ctx context.Context, userID,
	resumeID uuid.UUID, jobText string, projectControls []ProjectControl, disciplineOverride *Discipline) (Run, error)`

- File: `service.go`
- Purpose: Creates a new record or resource and returns the result.

### `func (s *Service) GetRunByID(ctx context.Context, userID, runID uuid.UUID) (Run, error)`

- File: `service.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func (s *Service) ListRunsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Run, error)`

- File: `service.go`
- Purpose: Fetches a paginated or ordered list of resources.

### `func NewService(repo *Repo, jobsEnq jobs.JobsEnqueuer) *Service`

- File: `service.go`
- Purpose: Creates and returns a new instance used by this package.

### `func (s *Service) UpdateRunDiscipline(ctx context.Context, runID uuid.UUID, discipline Discipline, confidence float64, source DisciplineSource) error`

- File: `service.go`
- Purpose: Updates persisted state for an existing resource.

### `func normalizeProjectControls(controls []ProjectControl) []ProjectControl`

- File: `service.go`
- Purpose: Normalizes text or structures before scoring/persistence.

### `func ParseDiscipline(raw string) (Discipline, bool)`

- File: `types.go`
- Purpose: Parses input text into a validated structured value.

