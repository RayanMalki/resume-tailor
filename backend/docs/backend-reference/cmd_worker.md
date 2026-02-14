# cmd/worker

- Relative path: `cmd/worker`
- Packages: `main`
- Source files: `main.go`

## Functions

### `func (a *runsRepoAdapter) GetRunByID(ctx context.Context, runID uuid.UUID) (jobs.RunData, error)`

- File: `main.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func (a *runsRepoAdapter) UpdateRunDiscipline(ctx context.Context, runID uuid.UUID, discipline string, confidence float64, source string) error`

- File: `main.go`
- Purpose: Updates persisted state for an existing resource.

### `func main()`

- File: `main.go`
- Purpose: Internal function supporting `cmd/worker` package workflows.

### `func mapProjectControls(controls []runs.ProjectControl) []ai.ProjectControl`

- File: `main.go`
- Purpose: Internal function supporting `cmd/worker` package workflows.

