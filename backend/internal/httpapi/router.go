package httpapi

import (
	"net/http"
	"resume-tailor/internal/httpapi/handlers"
)

func NewRouter() http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/health", handlers.HandleHealth)

	return mux
}
