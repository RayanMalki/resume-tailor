package main

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"sort"
	"strings"

	"resume-tailor/internal/config"
	"resume-tailor/internal/db"
	"resume-tailor/migrations"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to db", "error", err)
		os.Exit(1)
	}
	defer db.Close(pool)

	files, err := listMigrationFiles()
	if err != nil {
		slog.Error("failed to list migrations", "error", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		slog.Info("no migrations found")
		return
	}

	for _, name := range files {
		sqlBytes, err := fs.ReadFile(migrations.FS, name)
		if err != nil {
			slog.Error("failed to read migration", "file", name, "error", err)
			os.Exit(1)
		}

		sql := strings.TrimSpace(string(sqlBytes))
		if sql == "" {
			continue
		}

		if _, err := pool.Exec(ctx, sql); err != nil {
			slog.Error("migration failed", "file", name, "error", err)
			os.Exit(1)
		}

		slog.Info("migration applied", "file", name)
	}

	slog.Info("migrations complete")
}

func listMigrationFiles() ([]string, error) {
	entries, err := fs.ReadDir(migrations.FS, ".")
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
