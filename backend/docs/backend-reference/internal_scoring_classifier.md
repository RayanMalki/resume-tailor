# internal/scoring/classifier

- Relative path: `internal/scoring/classifier`
- Packages: `classifier`
- Source files: `classifier.go`

## Functions

### `func Detect(resumeText, jobText string) Detection`

- File: `classifier.go`
- Purpose: Internal function supporting `internal/scoring/classifier` package workflows.

### `func Resolve(resumeText, jobText string, override *profiles.Discipline) Detection`

- File: `classifier.go`
- Purpose: Internal function supporting `internal/scoring/classifier` package workflows.

### `func detectWithThreshold(resumeText, jobText string, threshold float64) Detection`

- File: `classifier.go`
- Purpose: Internal function supporting `internal/scoring/classifier` package workflows.

### `func freq(tokens []string) map[string]int`

- File: `classifier.go`
- Purpose: Internal function supporting `internal/scoring/classifier` package workflows.

### `func normalizeText(text string) string`

- File: `classifier.go`
- Purpose: Normalizes text or structures before scoring/persistence.

### `func prefix(s string, n int) string`

- File: `classifier.go`
- Purpose: Internal function supporting `internal/scoring/classifier` package workflows.

### `func scoreProfile(profile profiles.Profile, tokenFreq map[string]int, jobLead, resumeLead string) (float64, []EvidenceTerm)`

- File: `classifier.go`
- Purpose: Internal function supporting `internal/scoring/classifier` package workflows.

### `func stripAccents(s string) string`

- File: `classifier.go`
- Purpose: Internal function supporting `internal/scoring/classifier` package workflows.

### `func tokenize(text string) []string`

- File: `classifier.go`
- Purpose: Handles token creation, hashing, parsing, or token-derived state.

