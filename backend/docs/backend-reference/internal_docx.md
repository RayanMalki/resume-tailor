# internal/docx

- Relative path: `internal/docx`
- Packages: `docx`
- Source files: `render.go`

## Functions

### `func RenderResume(spec ai.ResumeSpec) ([]byte, error)`

- File: `render.go`
- Purpose: RenderResume builds a minimal .docx file from the tailored resume spec.

### `func buildDocumentXML(paragraphs []string) string`

- File: `render.go`
- Purpose: Builds derived output text/data from normalized inputs.

### `func resumeParagraphs(spec ai.ResumeSpec) []string`

- File: `render.go`
- Purpose: Internal function supporting `internal/docx` package workflows.

### `func xmlEscape(s string) string`

- File: `render.go`
- Purpose: Internal function supporting `internal/docx` package workflows.

