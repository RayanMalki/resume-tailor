# internal/httpapi

- Relative path: `internal/httpapi`
- Packages: `httpapi`
- Source files: `router.go`

## Functions

### `func NewRouter(authSvc *auth.Service, runsSvc *runs.Service, resumesSvc *resumes.Service, reportsSvc *runreports.Service, artifactsSvc *artifacts.Service, emailSvc *email.Sender, allowedOrigins []string) http.Handler`

- File: `router.go`
- Purpose: Creates and returns a new instance used by this package.

