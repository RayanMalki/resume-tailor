# internal/httpapi/cookies

- Relative path: `internal/httpapi/cookies`
- Packages: `cookies`
- Source files: `cookies.go`

## Functions

### `func ClearOAuthRedirectCookie(w http.ResponseWriter)`

- File: `cookies.go`
- Purpose: Applies authentication checks and request identity propagation.

### `func ClearOAuthStateCookie(w http.ResponseWriter)`

- File: `cookies.go`
- Purpose: Applies authentication checks and request identity propagation.

### `func ClearSessionCookie(w http.ResponseWriter)`

- File: `cookies.go`
- Purpose: Internal function supporting `internal/httpapi/cookies` package workflows.

### `func ReadOAuthRedirectCookie(r *http.Request) (string, bool)`

- File: `cookies.go`
- Purpose: Applies authentication checks and request identity propagation.

### `func ReadOAuthStateCookie(r *http.Request) (string, bool)`

- File: `cookies.go`
- Purpose: Applies authentication checks and request identity propagation.

### `func ReadSessionCookie(r *http.Request) (token string, ok bool)`

- File: `cookies.go`
- Purpose: Internal function supporting `internal/httpapi/cookies` package workflows.

### `func SetOAuthRedirectCookie(w http.ResponseWriter, path string, expiresAt time.Time)`

- File: `cookies.go`
- Purpose: Applies authentication checks and request identity propagation.

### `func SetOAuthStateCookie(w http.ResponseWriter, state string, expiresAt time.Time)`

- File: `cookies.go`
- Purpose: Applies authentication checks and request identity propagation.

### `func SetSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time)`

- File: `cookies.go`
- Purpose: Internal function supporting `internal/httpapi/cookies` package workflows.

### `func envTruthy(key string) bool`

- File: `cookies.go`
- Purpose: Internal function supporting `internal/httpapi/cookies` package workflows.

### `func parseSameSite(value string) http.SameSite`

- File: `cookies.go`
- Purpose: Internal function supporting `internal/httpapi/cookies` package workflows.

### `func sessionCookieOptions() (bool, http.SameSite, string)`

- File: `cookies.go`
- Purpose: Internal function supporting `internal/httpapi/cookies` package workflows.

