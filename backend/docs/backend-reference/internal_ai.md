# internal/ai

- Relative path: `internal/ai`
- Packages: `ai`
- Source files: `client.go`

## Functions

### `func (c *Client) GenerateCoverLetter(ctx context.Context, resumeText, jobText string, bm25Signals any, lang string, disciplineCtx DisciplineContext) (string, error)`

- File: `client.go`
- Purpose: Internal function supporting `internal/ai` package workflows.

### `func (c *Client) GenerateProjectReasons(ctx context.Context, resumeText, jobText, latex string, bm25Signals any, projectControls []ProjectControl) ([]ProjectReason, error)`

- File: `client.go`
- Purpose: Internal function supporting `internal/ai` package workflows.

### `func (c *Client) GenerateResumeLatex(ctx context.Context, resumeText, jobText string, bm25Signals any) (string, error)`

- File: `client.go`
- Purpose: GenerateResumeLatex generates a Jake's Resume-style LaTeX output tailored to the job.

### `func (c *Client) GenerateResumeSpec(ctx context.Context, resumeText, jobText string, bm25Signals any, projectControls []ProjectControl, disciplineCtx DisciplineContext) (ResumeSpec, error)`

- File: `client.go`
- Purpose: GenerateResumeSpec generates a strict JSON resume spec for a fixed template.

### `func (c *Client) GenerateRunReport(ctx context.Context, addedKeywords []string, missingTerms []string, resumeLang string, disciplineCtx DisciplineContext) (ATSReport, ChangePlan, error)`

- File: `client.go`
- Purpose: GenerateRunReport generates a short ATS summary and notes using OpenAI.

### `func NewClientFromEnv(apiKey, model string) (*Client, error)`

- File: `client.go`
- Purpose: NewClientFromEnv creates a new OpenAI client from environment variables.

### `func ResumeSpecToText(spec ResumeSpec) string`

- File: `client.go`
- Purpose: ResumeSpecToText converts a ResumeSpec to plain text for BM25 analysis.

### `func buildCoverLetterPrompt(resumeText, jobText string, bm25Signals any, lang string, disciplineCtx DisciplineContext) string`

- File: `client.go`
- Purpose: Builds derived output text/data from normalized inputs.

### `func buildProjectReasonsPrompt(resumeText, jobText, latex string, bm25Signals any, projectControls []ProjectControl) string`

- File: `client.go`
- Purpose: Builds derived output text/data from normalized inputs.

### `func buildReportPrompt(addedKeywords, missingTerms []string, lang string, disciplineCtx DisciplineContext) string`

- File: `client.go`
- Purpose: Builds derived output text/data from normalized inputs.

### `func buildResumeLatexPrompt(resumeText, jobText string, bm25Signals any) string`

- File: `client.go`
- Purpose: Builds derived output text/data from normalized inputs.

### `func buildResumeSpecPrompt(resumeText, jobText string, bm25Signals any, projectControls []ProjectControl, disciplineCtx DisciplineContext) string`

- File: `client.go`
- Purpose: Builds derived output text/data from normalized inputs.

### `func closeOpenBraces(s string) (string, bool)`

- File: `client.go`
- Purpose: Internal function supporting `internal/ai` package workflows.

### `func unmarshalJSONObjectWithRecovery(content string, out any) error`

- File: `client.go`
- Purpose: Internal function supporting `internal/ai` package workflows.

