package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"resume-tailor/internal/config"
	"resume-tailor/internal/db"
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

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/health", handleHealth)

	fmt.Printf("listening to: %v", cfg.HTTPAddr)
	http.ListenAndServe(cfg.HTTPAddr, mux)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello")
}
func handleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}
