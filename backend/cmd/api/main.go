package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"resume-tailor/internal/artifacts"
	"resume-tailor/internal/auth"
	"resume-tailor/internal/config"
	"resume-tailor/internal/db"
	"resume-tailor/internal/email"
	"resume-tailor/internal/httpapi"
	"resume-tailor/internal/jobs"
	"resume-tailor/internal/monitoring"
	"resume-tailor/internal/resumes"
	"resume-tailor/internal/runreports"
	"resume-tailor/internal/runs"
	"resume-tailor/migrations"
)

func main() {
	ctx := context.Background()

	// Initialize monitoring (Sentry when SENTRY_DSN is set).
	monitoring.Init("resume-tailor-api")
	defer monitoring.Flush(2 * time.Second)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	if strings.TrimSpace(os.Getenv("DISCORD_WEBHOOK_URL")) == "" {
		slog.Warn("discord webhook disabled (DISCORD_WEBHOOK_URL not set)")
	} else {
		slog.Info("discord webhook enabled")
	}
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to db", "error", err)
		os.Exit(1)
	}
	defer db.Close(pool)

	slog.Info("running migrations")
	if err := migrations.Run(ctx, pool); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	authRepo := auth.NewRepo(pool)
	authSvc := auth.NewService(authRepo)
	jobsRepo := jobs.NewRepo(pool)
	runsRepo := runs.NewRepo(pool)
	runsSvc := runs.NewService(runsRepo, jobsRepo)
	resumesRepo := resumes.NewRepo(pool)
	resumesSvc := resumes.NewService(resumesRepo)
	runreportsRepo := runreports.NewRepo(pool)
	runreportsSvc := runreports.NewService(runreportsRepo)
	artifactsRepo := artifacts.NewRepo(pool)
	artifactsSvc := artifacts.NewService(artifactsRepo)

	allowedOrigins := []string{}
	for _, origin := range strings.Split(cfg.FrontendOrigin, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowedOrigins = append(allowedOrigins, origin)
		}
	}

	// Email service
	emailSvc := email.NewSender()

	// Start background session cleanup (runs every hour, stops on shutdown).
	cleanupCtx, cleanupCancel := context.WithCancel(ctx)
	defer cleanupCancel()
	authSvc.StartSessionCleanup(cleanupCtx, 1*time.Hour)

	router := httpapi.NewRouter(authSvc, runsSvc, resumesSvc, runreportsSvc, artifactsSvc, emailSvc, allowedOrigins)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		slog.Info("server starting", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Listen for SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
