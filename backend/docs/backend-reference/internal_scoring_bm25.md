# internal/scoring/bm25

- Relative path: `internal/scoring/bm25`
- Packages: `bm25`
- Source files: `bm25.go`, `idf_table.go`, `stopwords.go`

## Functions

### `func Compute(resumeText, jobText string) (Signals, error)`

- File: `bm25.go`
- Purpose: Compute calculates BM25 signals for resume and job text matching.

### `func ComputeWithProfile(resumeText, jobText string, profile profiles.Profile) (Signals, error)`

- File: `bm25.go`
- Purpose: Internal function supporting `internal/scoring/bm25` package workflows.

### `func bucketForTerm(profile profiles.Profile, term string) string`

- File: `bm25.go`
- Purpose: Internal function supporting `internal/scoring/bm25` package workflows.

### `func canonicalize(token string) string`

- File: `bm25.go`
- Purpose: Internal function supporting `internal/scoring/bm25` package workflows.

### `func canonicalizeWithProfile(token string, profile profiles.Profile) string`

- File: `bm25.go`
- Purpose: Internal function supporting `internal/scoring/bm25` package workflows.

### `func classifyTerm(term string) string`

- File: `bm25.go`
- Purpose: Internal function supporting `internal/scoring/bm25` package workflows.

### `func depluralize(token string) string`

- File: `bm25.go`
- Purpose: Depluralize applies simple plural→singular normalization for English and French.

### `func isDigitsOnly(term string) bool`

- File: `bm25.go`
- Purpose: Returns a boolean predicate used by surrounding workflow guards.

### `func isLowSignalTerm(term, category string, idf float64, profile profiles.Profile) bool`

- File: `bm25.go`
- Purpose: Returns a boolean predicate used by surrounding workflow guards.

### `func normalizeForTokenization(text string) string`

- File: `bm25.go`
- Purpose: Normalizes text or structures before scoring/persistence.

### `func sortStrings(values []string)`

- File: `bm25.go`
- Purpose: Internal helper for deterministic normalization/ordering.

### `func sortTermScores(items []TermScore)`

- File: `bm25.go`
- Purpose: Internal helper for deterministic normalization/ordering.

### `func stripAccents(s string) string`

- File: `bm25.go`
- Purpose: StripAccents removes diacritics/accents from text using Unicode NFD.

### `func termFreq(tokens []string) map[string]int`

- File: `bm25.go`
- Purpose: Internal function supporting `internal/scoring/bm25` package workflows.

### `func tokenize(text string) []string`

- File: `bm25.go`
- Purpose: Handles token creation, hashing, parsing, or token-derived state.

### `func tokenizeWithProfile(text string, profile profiles.Profile) []string`

- File: `bm25.go`
- Purpose: Handles token creation, hashing, parsing, or token-derived state.

### `func isLikelyNoisyToken(term string) bool`

- File: `idf_table.go`
- Purpose: Handles token creation, hashing, parsing, or token-derived state.

### `func lookupIDF(term string) float64`

- File: `idf_table.go`
- Purpose: LookupIDF returns the IDF value for a given term from the static table.

### `func addWords(block string)`

- File: `stopwords.go`
- Purpose: Internal function supporting `internal/scoring/bm25` package workflows.

### `func init()`

- File: `stopwords.go`
- Purpose: Internal function supporting `internal/scoring/bm25` package workflows.

### `func isStopword(token string) bool`

- File: `stopwords.go`
- Purpose: Returns a boolean predicate used by surrounding workflow guards.

