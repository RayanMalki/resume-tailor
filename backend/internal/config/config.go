package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL  string
	HTTPAddr     string
	WorkerID     string
	OpenAIAPIKey string
	OpenAIModel  string
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		HTTPAddr:     os.Getenv("HTTP_ADDR"),
		WorkerID:     os.Getenv("WORKER_ID"),
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:  os.Getenv("OPENAI_MODEL"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8080"
	}

	if cfg.WorkerID == "" {
		cfg.WorkerID = "worker-1"
	}

	if cfg.OpenAIModel == "" {
		cfg.OpenAIModel = "gpt-4o-mini"
	}

	return cfg, nil
}
