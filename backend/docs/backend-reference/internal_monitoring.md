# internal/monitoring

- Relative path: `internal/monitoring`
- Packages: `monitoring`
- Source files: `sentry.go`

## Functions

### `func CaptureError(err error, context map[string]string)`

- File: `sentry.go`
- Purpose: CaptureError logs an error and sends it to Sentry with optional tags.

### `func CaptureMessage(msg string, context map[string]string)`

- File: `sentry.go`
- Purpose: CaptureMessage logs an informational message and sends it to Sentry.

### `func (e *sentryError) Error() string`

- File: `sentry.go`
- Purpose: Internal function supporting `internal/monitoring` package workflows.

### `func Flush(timeout time.Duration)`

- File: `sentry.go`
- Purpose: Flush should be called before application shutdown to ensure all events.

### `func Init(service string)`

- File: `sentry.go`
- Purpose: Init initializes monitoring. Call this early in main().

### `func RecoverMiddleware(next http.Handler) http.Handler`

- File: `sentry.go`
- Purpose: RecoverMiddleware catches panics in HTTP handlers, logs them, reports to.

### `func RequestMetricsMiddleware(next http.Handler) http.Handler`

- File: `sentry.go`
- Purpose: RequestMetricsMiddleware logs request duration and status for observability.

### `func TrackMetric(event string, attrs map[string]string)`

- File: `sentry.go`
- Purpose: TrackMetric logs a key metric event.

### `func (w *statusWriter) WriteHeader(code int)`

- File: `sentry.go`
- Purpose: Internal function supporting `internal/monitoring` package workflows.

### `func buildVersion() string`

- File: `sentry.go`
- Purpose: BuildVersion returns the VCS revision from build info, or "unknown".

### `func toError(v interface{}) error`

- File: `sentry.go`
- Purpose: ToError converts a recovered panic value to an error.

