# migrations

- Relative path: `migrations`
- Packages: `migrations`
- Source files: `embed.go`, `migrate.go`

## Functions

### `func Run(ctx context.Context, pool *pgxpool.Pool) error`

- File: `migrate.go`
- Purpose: Run applies all embedded SQL migrations in filename order.

### `func applyEnumMigrationInAutocommit(ctx context.Context, conn *pgx.Conn, sql string) error`

- File: `migrate.go`
- Purpose: Internal function supporting `migrations` package workflows.

### `func ensureArtifactsItemsTable(ctx context.Context, pool *pgxpool.Pool) error`

- File: `migrate.go`
- Purpose: Internal function supporting `migrations` package workflows.

### `func ensureMigrationState(ctx context.Context, pool *pgxpool.Pool) error`

- File: `migrate.go`
- Purpose: Internal function supporting `migrations` package workflows.

### `func extractGooseUpSQL(content string) string`

- File: `migrate.go`
- Purpose: Extracts structured text/data from raw uploaded content.

### `func isApplied(ctx context.Context, conn *pgx.Conn, name string) (bool, error)`

- File: `migrate.go`
- Purpose: Returns a boolean predicate used by surrounding workflow guards.

### `func isUnsafeEnumUse(err error) bool`

- File: `migrate.go`
- Purpose: Returns a boolean predicate used by surrounding workflow guards.

### `func listFiles() ([]string, error)`

- File: `migrate.go`
- Purpose: Internal function supporting `migrations` package workflows.

### `func markApplied(ctx context.Context, conn *pgx.Conn, name string) error`

- File: `migrate.go`
- Purpose: Internal function supporting `migrations` package workflows.

### `func readDollarTag(sql string) string`

- File: `migrate.go`
- Purpose: Internal function supporting `migrations` package workflows.

### `func splitSQLStatements(sql string) []string`

- File: `migrate.go`
- Purpose: Internal helper for deterministic normalization/ordering.

