package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL      string
	HTTPAddr         string
	WorkerID         string
	WorkerJobTimeout time.Duration
	DisciplineMode   string
	OpenAIAPIKey     string
	OpenAIModel      string
	FrontendOrigin   string
	PDFEnabled       bool
	TectonicBin      string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		HTTPAddr:         os.Getenv("HTTP_ADDR"),
		WorkerID:         os.Getenv("WORKER_ID"),
		WorkerJobTimeout: 15 * time.Minute,
		DisciplineMode:   os.Getenv("DISCIPLINE_MODE"),
		OpenAIAPIKey:     os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:      os.Getenv("OPENAI_MODEL"),
		FrontendOrigin:   os.Getenv("FRONTEND_ORIGIN"),
		PDFEnabled:       os.Getenv("RESUME_PDF_ENABLED") == "1",
		TectonicBin:      os.Getenv("TECTONIC_BIN"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.HTTPAddr == "" {
		if port := os.Getenv("PORT"); port != "" {
			cfg.HTTPAddr = ":" + port
		} else {
			cfg.HTTPAddr = ":8080"
		}
	}

	if cfg.WorkerID == "" {
		cfg.WorkerID = "worker-1"
	}
	if raw := os.Getenv("WORKER_JOB_TIMEOUT"); raw != "" {
		dur, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, fmt.Errorf("invalid WORKER_JOB_TIMEOUT %q: %w", raw, err)
		}
		if dur <= 0 {
			return Config{}, fmt.Errorf("WORKER_JOB_TIMEOUT must be > 0")
		}
		cfg.WorkerJobTimeout = dur
	}

	if cfg.OpenAIModel == "" {
		cfg.OpenAIModel = "gpt-4o-mini"
	}

	switch cfg.DisciplineMode {
	case "", "enforce":
		cfg.DisciplineMode = "enforce"
	case "off", "observe":
		// valid values
	default:
		return Config{}, fmt.Errorf("invalid DISCIPLINE_MODE %q: must be off|observe|enforce", cfg.DisciplineMode)
	}

	if cfg.FrontendOrigin == "" {
		cfg.FrontendOrigin = "http://localhost:3000"
	}

	if cfg.TectonicBin == "" {
		cfg.TectonicBin = "tectonic"
	}

	return cfg, nil
}
