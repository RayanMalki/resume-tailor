package migrations

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Run applies all embedded SQL migrations in filename order.
func Run(ctx context.Context, pool *pgxpool.Pool) error {
	if err := ensureMigrationState(ctx, pool); err != nil {
		return err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration conn: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", int64(982451653)); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		_, _ = conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", int64(982451653))
	}()

	files, err := listFiles()
	if err != nil {
		return err
	}

	for _, name := range files {
		applied, err := isApplied(ctx, conn.Conn(), name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		sqlBytes, err := fs.ReadFile(FS, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		sql := strings.TrimSpace(extractGooseUpSQL(string(sqlBytes)))
		if sql == "" {
			continue
		}

		tx, err := conn.Conn().Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}

		if _, err := tx.Exec(ctx, sql); err != nil {
			_ = tx.Rollback(ctx)
			if isUnsafeEnumUse(err) {
				if err := applyEnumMigrationInAutocommit(ctx, conn.Conn(), sql); err != nil {
					return fmt.Errorf("apply migration %s: %w", name, err)
				}
				if err := markApplied(ctx, conn.Conn(), name); err != nil {
					return err
				}
				continue
			}
			return fmt.Errorf("apply migration %s: %w", name, err)
		}

		if _, err := tx.Exec(ctx, `
INSERT INTO schema_migrations (name, applied_at)
VALUES ($1, now())
ON CONFLICT (name) DO NOTHING
`, name); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", name, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}

	if err := ensureArtifactsItemsTable(ctx, pool); err != nil {
		return fmt.Errorf("ensure run_artifacts_items: %w", err)
	}

	return nil
}

func extractGooseUpSQL(content string) string {
	upMarker := "-- +goose Up"
	downMarker := "-- +goose Down"

	upIdx := strings.Index(content, upMarker)
	if upIdx == -1 {
		return content
	}

	afterUp := content[upIdx+len(upMarker):]
	downIdx := strings.Index(afterUp, downMarker)
	if downIdx == -1 {
		return afterUp
	}
	return afterUp[:downIdx]
}

func listFiles() ([]string, error) {
	entries, err := fs.ReadDir(FS, ".")
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".sql") {
			files = append(files, name)
		}
	}

	sort.Strings(files)
	return files, nil
}

func isUnsafeEnumUse(err error) bool {
	return strings.Contains(err.Error(), "unsafe use of new value")
}

func applyEnumMigrationInAutocommit(ctx context.Context, conn *pgx.Conn, sql string) error {
	for _, stmt := range splitSQLStatements(sql) {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := conn.Exec(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func splitSQLStatements(sql string) []string {
	var stmts []string
	var buf strings.Builder
	inSingle := false
	inDouble := false
	inDollar := false
	dollarTag := ""

	for i := 0; i < len(sql); i++ {
		ch := sql[i]

		if inDollar {
			if ch == '$' && strings.HasPrefix(sql[i:], dollarTag+"$") {
				inDollar = false
				buf.WriteString(dollarTag)
				buf.WriteByte('$')
				i += len(dollarTag)
				continue
			}
			buf.WriteByte(ch)
			continue
		}

		if !inSingle && !inDouble && ch == '$' {
			tag := readDollarTag(sql[i:])
			if tag != "" {
				inDollar = true
				dollarTag = tag
				buf.WriteString(tag)
				buf.WriteByte('$')
				i += len(tag)
				continue
			}
		}

		if ch == '\'' && !inDouble {
			inSingle = !inSingle
		} else if ch == '"' && !inSingle {
			inDouble = !inDouble
		}

		if ch == ';' && !inSingle && !inDouble && !inDollar {
			stmts = append(stmts, buf.String())
			buf.Reset()
			continue
		}

		buf.WriteByte(ch)
	}

	if strings.TrimSpace(buf.String()) != "" {
		stmts = append(stmts, buf.String())
	}

	return stmts
}

func readDollarTag(sql string) string {
	if len(sql) < 2 || sql[0] != '$' {
		return ""
	}
	end := strings.IndexByte(sql[1:], '$')
	if end == -1 {
		return ""
	}
	return sql[:end+1]
}

func ensureArtifactsItemsTable(ctx context.Context, pool *pgxpool.Pool) error {
	const q = `
CREATE TABLE IF NOT EXISTS run_artifacts_items (
  run_id    UUID NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  type      TEXT NOT NULL,
  content   TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (run_id, type)
);

CREATE INDEX IF NOT EXISTS idx_run_artifacts_items_run_id
  ON run_artifacts_items(run_id);
`
	_, err := pool.Exec(ctx, q)
	return err
}

func ensureMigrationState(ctx context.Context, pool *pgxpool.Pool) error {
	const q = `
CREATE TABLE IF NOT EXISTS schema_migrations (
  name TEXT PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);`
	if _, err := pool.Exec(ctx, q); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}
	return nil
}

func isApplied(ctx context.Context, conn *pgx.Conn, name string) (bool, error) {
	const q = `SELECT 1 FROM schema_migrations WHERE name = $1`
	var one int
	err := conn.QueryRow(ctx, q, name).Scan(&one)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("check migration state %s: %w", name, err)
	}
	return true, nil
}

func markApplied(ctx context.Context, conn *pgx.Conn, name string) error {
	const q = `
INSERT INTO schema_migrations (name, applied_at)
VALUES ($1, now())
ON CONFLICT (name) DO NOTHING`
	if _, err := conn.Exec(ctx, q, name); err != nil {
		return fmt.Errorf("record migration %s: %w", name, err)
	}
	return nil
}
