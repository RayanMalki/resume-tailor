package httpapi

import (
	"net/http"
	"resume-tailor/internal/httpapi/handlers"
	"resume-tailor/internal/httpapi/middleware"

	"github.com/go-chi/chi/v5"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/", handlers.RootHandler)
	r.Get("/health", handlers.HandleHealth)

	// v1 API routes
	r.Route("/v1", func(r chi.Router) {
		// Add v1 routes here in the future
	})

	// NotFound handler returns JSON 404
	r.NotFound(handlers.HandleNotFound)

	handler := middleware.Recover(middleware.Logging(r))
	return handler
}
