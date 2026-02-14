# internal/artifacts

- Relative path: `internal/artifacts`
- Packages: `artifacts`
- Source files: `repo.go`, `service.go`, `types.go`

## Functions

### `func (r *Repo) GetByRunIDAndType(ctx context.Context, runID uuid.UUID, artifactType string) (Artifact, error)`

- File: `repo.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func (r *Repo) InsertIfNotExists(ctx context.Context, runID uuid.UUID, artifactType string, content string) error`

- File: `repo.go`
- Purpose: Internal function supporting `internal/artifacts` package workflows.

### `func NewRepo(db *pgxpool.Pool) *Repo`

- File: `repo.go`
- Purpose: Creates and returns a new instance used by this package.

### `func (s *Service) GetByRunIDAndType(ctx context.Context, runID uuid.UUID, artifactType string) (Artifact, error)`

- File: `service.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func (s *Service) InsertIfNotExists(ctx context.Context, runID uuid.UUID, artifactType string, content string) error`

- File: `service.go`
- Purpose: Internal function supporting `internal/artifacts` package workflows.

### `func NewService(repo *Repo) *Service`

- File: `service.go`
- Purpose: Creates and returns a new instance used by this package.

