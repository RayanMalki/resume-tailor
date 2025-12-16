package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"resume-tailor/internal/config"
	"resume-tailor/internal/db"
	"resume-tailor/internal/httpapi"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to log config: %v", err)
	}
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close(pool)

	router := httpapi.NewRouter()

	fmt.Printf("listening to: %v", cfg.HTTPAddr)
	err = http.ListenAndServe(cfg.HTTPAddr, router)
	if err != nil {
		log.Fatalf("server error: %v", err)
	}
}
