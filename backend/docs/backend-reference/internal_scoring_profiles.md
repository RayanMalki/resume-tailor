# internal/scoring/profiles

- Relative path: `internal/scoring/profiles`
- Packages: `profiles`
- Source files: `profiles.go`, `types.go`

## Functions

### `func BucketForTerm(profile Profile, term string) string`

- File: `profiles.go`
- Purpose: Internal function supporting `internal/scoring/profiles` package workflows.

### `func BucketWeight(profile Profile, bucket string) float64`

- File: `profiles.go`
- Purpose: Internal function supporting `internal/scoring/profiles` package workflows.

### `func Canonicalize(profile Profile, token string) string`

- File: `profiles.go`
- Purpose: Internal function supporting `internal/scoring/profiles` package workflows.

### `func DefaultDiscipline() Discipline`

- File: `profiles.go`
- Purpose: Internal function supporting `internal/scoring/profiles` package workflows.

### `func Get(d Discipline) Profile`

- File: `profiles.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func GetAll() map[Discipline]Profile`

- File: `profiles.go`
- Purpose: Fetches a single resource and returns it with validation/error handling.

### `func IsLowSignal(profile Profile, token string) bool`

- File: `profiles.go`
- Purpose: Internal function supporting `internal/scoring/profiles` package workflows.

### `func Version() string`

- File: `profiles.go`
- Purpose: Internal function supporting `internal/scoring/profiles` package workflows.

### `func cloneBoolMap(in map[string]bool) map[string]bool`

- File: `profiles.go`
- Purpose: Internal function supporting `internal/scoring/profiles` package workflows.

### `func cloneBuckets(in map[string][]string) map[string][]string`

- File: `profiles.go`
- Purpose: Internal function supporting `internal/scoring/profiles` package workflows.

### `func cloneFloatMap(in map[string]float64) map[string]float64`

- File: `profiles.go`
- Purpose: Internal function supporting `internal/scoring/profiles` package workflows.

### `func cloneProfile(in Profile) Profile`

- File: `profiles.go`
- Purpose: Internal function supporting `internal/scoring/profiles` package workflows.

### `func cloneStringMap(in map[string]string) map[string]string`

- File: `profiles.go`
- Purpose: Internal function supporting `internal/scoring/profiles` package workflows.

### `func cloneStringSlice(in []string) []string`

- File: `profiles.go`
- Purpose: Internal function supporting `internal/scoring/profiles` package workflows.

### `func loadProfiles() map[Discipline]Profile`

- File: `profiles.go`
- Purpose: Internal function supporting `internal/scoring/profiles` package workflows.

### `func ParseDiscipline(raw string) (Discipline, bool)`

- File: `types.go`
- Purpose: Parses input text into a validated structured value.

### `func ValidDisciplines() []Discipline`

- File: `types.go`
- Purpose: Internal function supporting `internal/scoring/profiles` package workflows.

