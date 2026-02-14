# internal/latex

- Relative path: `internal/latex`
- Packages: `latex`
- Source files: `compile.go`, `template.go`

## Functions

### `func CompilePDF(ctx context.Context, tectonicBin string, latex string) ([]byte, error)`

- File: `compile.go`
- Purpose: Compiles generated source into a final binary/document artifact.

### `func RenderResume(spec ai.ResumeSpec) string`

- File: `template.go`
- Purpose: Renders final artifact content from structured inputs.

### `func clampAndEscape(items []string, max int) []string`

- File: `template.go`
- Purpose: Internal helper for deterministic normalization/ordering.

### `func clampAndEscapeBold(items []string, max int) []string`

- File: `template.go`
- Purpose: Internal helper for deterministic normalization/ordering.

### `func clampAndEscapeWith(items []string, max int, preserveBold bool) []string`

- File: `template.go`
- Purpose: Internal helper for deterministic normalization/ordering.

### `func clampEducation(items []ai.ResumeEducation, maxEntries int) []ai.ResumeEducation`

- File: `template.go`
- Purpose: Internal helper for deterministic normalization/ordering.

### `func clampExperience(items []ai.ResumeExperience, maxEntries, maxBullets int) []ai.ResumeExperience`

- File: `template.go`
- Purpose: Internal helper for deterministic normalization/ordering.

### `func clampProjects(items []ai.ResumeProject, maxBullets int) []ai.ResumeProject`

- File: `template.go`
- Purpose: Internal helper for deterministic normalization/ordering.

### `func clampSkillGroups(items []ai.ResumeSkillGroup, maxGroups, maxSkills int) []ai.ResumeSkillGroup`

- File: `template.go`
- Purpose: Internal helper for deterministic normalization/ordering.

### `func escapeLatex(input string) string`

- File: `template.go`
- Purpose: Internal function supporting `internal/latex` package workflows.

### `func escapeLatexCore(input string, preserveBold bool) string`

- File: `template.go`
- Purpose: Internal function supporting `internal/latex` package workflows.

### `func escapeLatexRaw(input string) string`

- File: `template.go`
- Purpose: Internal function supporting `internal/latex` package workflows.

### `func escapeLatexWithBold(input string) string`

- File: `template.go`
- Purpose: EscapeLatexWithBold escapes LaTeX special characters while converting.

### `func fallback(value, alt string) string`

- File: `template.go`
- Purpose: Fallback/safety helper used to keep processing resilient.

### `func isFrenchLanguage(language string) bool`

- File: `template.go`
- Purpose: Returns a boolean predicate used by surrounding workflow guards.

### `func joinAndEscape(items []string, sep string) string`

- File: `template.go`
- Purpose: Internal function supporting `internal/latex` package workflows.

### `func sectionLabels(language string) sectionSet`

- File: `template.go`
- Purpose: Internal function supporting `internal/latex` package workflows.

