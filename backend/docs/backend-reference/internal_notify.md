# internal/notify

- Relative path: `internal/notify`
- Packages: `notify`
- Source files: `discord.go`

## Functions

### `func SendEvent(ctx context.Context, e Event) error`

- File: `discord.go`
- Purpose: Internal function supporting `internal/notify` package workflows.

### `func cleanOneLine(s string) string`

- File: `discord.go`
- Purpose: Internal function supporting `internal/notify` package workflows.

### `func fallback(value, alt string) string`

- File: `discord.go`
- Purpose: Fallback/safety helper used to keep processing resilient.

### `func formatEvent(e Event) string`

- File: `discord.go`
- Purpose: Internal function supporting `internal/notify` package workflows.

### `func parseDurationMsEnv(key string, defMs int) time.Duration`

- File: `discord.go`
- Purpose: Internal function supporting `internal/notify` package workflows.

### `func parseRetryAfter(value string) time.Duration`

- File: `discord.go`
- Purpose: Internal function supporting `internal/notify` package workflows.

### `func redactIP(ip string) string`

- File: `discord.go`
- Purpose: Internal function supporting `internal/notify` package workflows.

### `func sendWithRetry(ctx context.Context, client *http.Client, baseReq *http.Request, body []byte, maxRetryAfter time.Duration) error`

- File: `discord.go`
- Purpose: Internal function supporting `internal/notify` package workflows.

### `func sleepCtx(ctx context.Context, d time.Duration) bool`

- File: `discord.go`
- Purpose: Internal function supporting `internal/notify` package workflows.

### `func startDispatcher()`

- File: `discord.go`
- Purpose: Internal function supporting `internal/notify` package workflows.

### `func truncate(s string, max int) string`

- File: `discord.go`
- Purpose: Internal function supporting `internal/notify` package workflows.

