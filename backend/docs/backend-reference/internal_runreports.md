# internal/runreports

- Relative path: `internal/runreports`
- Packages: `runreports`
- Source files: `repo.go`, `service.go`, `types.go`

## Functions

### `func (r *Repo) GetRunReportByRunID(ctx context.Context, runID uuid.UUID) (RunReport, error)`

- File: `repo.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func NewRepo(db *pgxpool.Pool) *Repo`

- File: `repo.go`
- Purpose: Creates and returns a new instance used by this package.

### `func (r *Repo) UpsertRunReport(ctx context.Context, runID uuid.UUID, atsReport, changePlan json.RawMessage) error`

- File: `repo.go`
- Purpose: Internal function supporting `internal/runreports` package workflows.

### `func (s *Service) GetRunReportByRunID(ctx context.Context, runID uuid.UUID) (RunReport, error)`

- File: `service.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func NewService(repo *Repo) *Service`

- File: `service.go`
- Purpose: Creates and returns a new instance used by this package.

### `func (s *Service) UpsertRunReport(ctx context.Context, runID uuid.UUID, atsReport, changePlan json.RawMessage) error`

- File: `service.go`
- Purpose: Internal function supporting `internal/runreports` package workflows.

