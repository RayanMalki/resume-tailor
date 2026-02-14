# internal/httpapi/handlers

- Relative path: `internal/httpapi/handlers`
- Packages: `handlers`
- Source files: `auth_google.go`, `auth_login.go`, `auth_logout.go`, `auth_signup.go`, `auth_verify.go`, `discipline_detect.go`, `extract_text.go`, `health.go`, `json.go`, `me.go`, `notfound.go`, `resumes_create.go`, `resumes_get.go`, `resumes_list.go`, `resumes_upload.go`, `root.go`, `run_artifact_cover_letter_get.go`, `run_artifact_docx_get.go`, `run_artifact_latex_get.go`, `run_artifact_pdf_get.go`, `run_artifact_project_reasons_get.go`, `run_report_get.go`, `runs_create.go`, `runs_get.go`, `runs_list.go`

## Functions

### `func GoogleCallback(authSvc *auth.Service) http.HandlerFunc`

- File: `auth_google.go`
- Purpose: Internal function supporting `internal/httpapi/handlers` package workflows.

### `func GoogleStart(authSvc *auth.Service) http.HandlerFunc`

- File: `auth_google.go`
- Purpose: Internal function supporting `internal/httpapi/handlers` package workflows.

### `func exchangeGoogleToken(ctx context.Context, clientID, clientSecret, redirectURL, code string) (googleTokenResponse, error)`

- File: `auth_google.go`
- Purpose: Handles token creation, hashing, parsing, or token-derived state.

### `func fetchGoogleUser(ctx context.Context, accessToken string) (googleUserInfo, error)`

- File: `auth_google.go`
- Purpose: Internal function supporting `internal/httpapi/handlers` package workflows.

### `func sanitizeRedirect(raw string) string`

- File: `auth_google.go`
- Purpose: Internal function supporting `internal/httpapi/handlers` package workflows.

### `func Login(authSvc *auth.Service) http.HandlerFunc`

- File: `auth_login.go`
- Purpose: Internal function supporting `internal/httpapi/handlers` package workflows.

### `func Logout(authSvc *auth.Service) http.HandlerFunc`

- File: `auth_logout.go`
- Purpose: Internal function supporting `internal/httpapi/handlers` package workflows.

### `func Signup(authSvc *auth.Service) http.HandlerFunc`

- File: `auth_signup.go`
- Purpose: Internal function supporting `internal/httpapi/handlers` package workflows.

### `func ForgotPassword(authSvc *auth.Service, emailSvc *email.Sender) http.HandlerFunc`

- File: `auth_verify.go`
- Purpose: ForgotPassword handles POST /v1/auth/forgot-password.

### `func ResendVerification(authSvc *auth.Service, emailSvc *email.Sender) http.HandlerFunc`

- File: `auth_verify.go`
- Purpose: ResendVerification handles POST /v1/auth/resend-verification.

### `func ResetPassword(authSvc *auth.Service) http.HandlerFunc`

- File: `auth_verify.go`
- Purpose: ResetPassword handles POST /v1/auth/reset-password.

### `func VerifyEmail(authSvc *auth.Service) http.HandlerFunc`

- File: `auth_verify.go`
- Purpose: VerifyEmail handles GET /v1/auth/verify?token=...

### `func DetectDisciplineHandler(resumesSvc *resumes.Service) http.HandlerFunc`

- File: `discipline_detect.go`
- Purpose: HTTP handler entry point that validates input and writes API responses.

### `func extractTextFromDOCX(data []byte) (string, error)`

- File: `extract_text.go`
- Purpose: Extracts structured text/data from raw uploaded content.

### `func extractTextFromPDF(data []byte) (string, error)`

- File: `extract_text.go`
- Purpose: Extracts structured text/data from raw uploaded content.

### `func parseDOCXXML(xmlContent []byte) (string, error)`

- File: `extract_text.go`
- Purpose: ParseDOCXXML extracts text from the Office Open XML document.

### `func HandleHealth(w http.ResponseWriter, r *http.Request)`

- File: `health.go`
- Purpose: HTTP handler entry point that validates input and writes API responses.

### `func decodeJSON(r *http.Request, v any) error`

- File: `json.go`
- Purpose: Decodes request or payload data into typed structures.

### `func primaryFrontendOrigin() string`

- File: `json.go`
- Purpose: PrimaryFrontendOrigin returns the first origin from the comma-separated.

### `func writeError(w http.ResponseWriter, status int, msg string)`

- File: `json.go`
- Purpose: Internal function supporting `internal/httpapi/handlers` package workflows.

### `func writeJSON(w http.ResponseWriter, status int, data any)`

- File: `json.go`
- Purpose: Internal function supporting `internal/httpapi/handlers` package workflows.

### `func Me() http.HandlerFunc`

- File: `me.go`
- Purpose: Internal function supporting `internal/httpapi/handlers` package workflows.

### `func HandleNotFound(w http.ResponseWriter, r *http.Request)`

- File: `notfound.go`
- Purpose: HTTP handler entry point that validates input and writes API responses.

### `func CreateResumeHandler(resumesSvc *resumes.Service) http.HandlerFunc`

- File: `resumes_create.go`
- Purpose: Creates a new record or resource and returns the result.

### `func GetResumeByIDHandler(resumesSvc *resumes.Service) http.HandlerFunc`

- File: `resumes_get.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func ListResumesHandler(resumesSvc *resumes.Service) http.HandlerFunc`

- File: `resumes_list.go`
- Purpose: Fetches a paginated or ordered list of resources.

### `func UploadResumeHandler(resumesSvc *resumes.Service) http.HandlerFunc`

- File: `resumes_upload.go`
- Purpose: UploadResumeHandler accepts a multipart/form-data file upload (PDF or DOCX),.

### `func RootHandler(w http.ResponseWriter, r *http.Request)`

- File: `root.go`
- Purpose: HTTP handler entry point that validates input and writes API responses.

### `func GetCoverLetterArtifactHandler(runsSvc *runs.Service, artifactsSvc *artifacts.Service) http.HandlerFunc`

- File: `run_artifact_cover_letter_get.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func GetResumeDOCXArtifactHandler(runsSvc *runs.Service, artifactsSvc *artifacts.Service) http.HandlerFunc`

- File: `run_artifact_docx_get.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func GetResumeLatexArtifactHandler(runsSvc *runs.Service, artifactsSvc *artifacts.Service) http.HandlerFunc`

- File: `run_artifact_latex_get.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func GetResumePDFArtifactHandler(runsSvc *runs.Service, artifactsSvc *artifacts.Service) http.HandlerFunc`

- File: `run_artifact_pdf_get.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func GetProjectReasonsArtifactHandler(runsSvc *runs.Service, artifactsSvc *artifacts.Service) http.HandlerFunc`

- File: `run_artifact_project_reasons_get.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func GetRunReportHandler(runsSvc *runs.Service, reportsSvc *runreports.Service) http.HandlerFunc`

- File: `run_report_get.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func CreateRunHandler(runsSvc *runs.Service, resumesSvc *resumes.Service) http.HandlerFunc`

- File: `runs_create.go`
- Purpose: Creates a new record or resource and returns the result.

### `func GetRunByIdHandler(runsSvc *runs.Service) http.HandlerFunc`

- File: `runs_get.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func ListRunsHandler(runsSvc *runs.Service) http.HandlerFunc`

- File: `runs_list.go`
- Purpose: Fetches a paginated or ordered list of resources.

