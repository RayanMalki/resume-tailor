# internal/resumes

- Relative path: `internal/resumes`
- Packages: `resumes`
- Source files: `repo.go`, `service.go`, `types.go`

## Functions

### `func (r *Repo) CreateResume(ctx context.Context, userID uuid.UUID, title string, contentText string) (Resume, error)`

- File: `repo.go`
- Purpose: Creates a new record or resource and returns the result.

### `func (r *Repo) GetResumeByID(ctx context.Context, resumeID uuid.UUID) (Resume, error)`

- File: `repo.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func (r *Repo) ListResumesByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Resume, error)`

- File: `repo.go`
- Purpose: Fetches a paginated or ordered list of resources.

### `func NewRepo(db *pgxpool.Pool) *Repo`

- File: `repo.go`
- Purpose: Creates and returns a new instance used by this package.

### `func (s *Service) CreateResume(ctx context.Context, userID uuid.UUID, title, contentText string) (Resume, error)`

- File: `service.go`
- Purpose: Creates a new record or resource and returns the result.

### `func (s *Service) GetResumeByID(ctx context.Context, userID, resumeID uuid.UUID) (Resume, error)`

- File: `service.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func (s *Service) ListResumesByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Resume, error)`

- File: `service.go`
- Purpose: Fetches a paginated or ordered list of resources.

### `func NewService(repo *Repo) *Service`

- File: `service.go`
- Purpose: Creates and returns a new instance used by this package.

