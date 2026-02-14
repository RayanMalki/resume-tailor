# internal/httpapi/middleware

- Relative path: `internal/httpapi/middleware`
- Packages: `middleware`
- Source files: `auth.go`, `auth_stub.go`, `cors.go`, `ip.go`, `limits.go`, `logging.go`, `ratelimit.go`, `recover.go`, `requestid.go`, `security.go`, `suspicious.go`

## Functions

### `func AuthRequired(authSvc *auth.Service) func(http.Handler) http.Handler`

- File: `auth.go`
- Purpose: AuthRequired is a middleware that validates the session cookie and injects the user ID into the context.

### `func UserIDFromContext(ctx context.Context) (uuid.UUID, bool)`

- File: `auth.go`
- Purpose: UserIDFromContext extracts the user ID from the request context.

### `func WithUserID(ctx context.Context, id uuid.UUID) context.Context`

- File: `auth.go`
- Purpose: WithUserID injects a user ID into the request context.

### `func writeUnauthorized(w http.ResponseWriter)`

- File: `auth.go`
- Purpose: WriteUnauthorized writes a 401 Unauthorized JSON response.

### `func CORS(allowedOrigins []string) func(http.Handler) http.Handler`

- File: `cors.go`
- Purpose: Internal function supporting `internal/httpapi/middleware` package workflows.

### `func ClientIP(r *http.Request) string`

- File: `ip.go`
- Purpose: Internal function supporting `internal/httpapi/middleware` package workflows.

### `func Limits(next http.Handler) http.Handler`

- File: `limits.go`
- Purpose: Applies request limiting or permit checks for abuse protection.

### `func matchBodyLimit(method, path string) int64`

- File: `limits.go`
- Purpose: Applies request limiting or permit checks for abuse protection.

### `func writeJSONError(w http.ResponseWriter, status int, msg string)`

- File: `limits.go`
- Purpose: Internal function supporting `internal/httpapi/middleware` package workflows.

### `func Logging(next http.Handler) http.Handler`

- File: `logging.go`
- Purpose: Internal function supporting `internal/httpapi/middleware` package workflows.

### `func (sr *statusRecorder) Write(b []byte) (int, error)`

- File: `logging.go`
- Purpose: Internal function supporting `internal/httpapi/middleware` package workflows.

### `func (sr *statusRecorder) WriteHeader(code int)`

- File: `logging.go`
- Purpose: Internal function supporting `internal/httpapi/middleware` package workflows.

### `func (l *Limiter) Allow(key string) (bool, RateLimitResult)`

- File: `ratelimit.go`
- Purpose: Applies request limiting or permit checks for abuse protection.

### `func AllowRunCreate(userID uuid.UUID) (bool, RateLimitResult)`

- File: `ratelimit.go`
- Purpose: Applies request limiting or permit checks for abuse protection.

### `func (rl *RateLimiter) AllowRunCreate(userID uuid.UUID) (bool, RateLimitResult)`

- File: `ratelimit.go`
- Purpose: Applies request limiting or permit checks for abuse protection.

### `func (rl *RateLimiter) Middleware(next http.Handler) http.Handler`

- File: `ratelimit.go`
- Purpose: Internal function supporting `internal/httpapi/middleware` package workflows.

### `func NewLimiter(limit int, window time.Duration) *Limiter`

- File: `ratelimit.go`
- Purpose: Creates and returns a new instance used by this package.

### `func NewRateLimiter(cfg RateLimitConfig) *RateLimiter`

- File: `ratelimit.go`
- Purpose: Creates and returns a new instance used by this package.

### `func RateLimit(next http.Handler) http.Handler`

- File: `ratelimit.go`
- Purpose: Applies request limiting or permit checks for abuse protection.

### `func (l *Limiter) cleanup(maxAge time.Duration)`

- File: `ratelimit.go`
- Purpose: Cleanup periodically removes buckets that haven't been accessed within maxAge.

### `func defaultRateLimitConfig() RateLimitConfig`

- File: `ratelimit.go`
- Purpose: Applies request limiting or permit checks for abuse protection.

### `func matchRoute(r *http.Request, prefix string) bool`

- File: `ratelimit.go`
- Purpose: Internal function supporting `internal/httpapi/middleware` package workflows.

### `func minFloat(a, b float64) float64`

- File: `ratelimit.go`
- Purpose: Internal function supporting `internal/httpapi/middleware` package workflows.

### `func (rl *RateLimiter) respondRateLimited(w http.ResponseWriter, r *http.Request, res RateLimitResult)`

- File: `ratelimit.go`
- Purpose: Applies request limiting or permit checks for abuse protection.

### `func Recover(next http.Handler) http.Handler`

- File: `recover.go`
- Purpose: Internal function supporting `internal/httpapi/middleware` package workflows.

### `func CSRFCheck(next http.Handler) http.Handler`

- File: `security.go`
- Purpose: CSRFCheck protects mutating endpoints against Cross-Site Request Forgery.

### `func SecurityHeaders(next http.Handler) http.Handler`

- File: `security.go`
- Purpose: SecurityHeaders adds standard security headers to every response.

### `func SuspiciousScanBlocker(next http.Handler) http.Handler`

- File: `suspicious.go`
- Purpose: Internal function supporting `internal/httpapi/middleware` package workflows.

### `func isSuspiciousPath(path string) bool`

- File: `suspicious.go`
- Purpose: Returns a boolean predicate used by surrounding workflow guards.

