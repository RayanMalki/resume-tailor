package httpapi

import (
	"net/http"

	"resume-tailor/internal/auth"
	"resume-tailor/internal/httpapi/handlers"
	"resume-tailor/internal/httpapi/middleware"

	"github.com/go-chi/chi/v5"
)

func NewRouter(authSvc *auth.Service) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Recover)
	r.Use(middleware.Logging)

	// v1 API routes
	r.Route("/v1", func(r chi.Router) {
		// Basic
		r.Get("/", handlers.RootHandler)
		r.Get("/health", handlers.HandleHealth)

		// Auth
		r.Post("/auth/signup", handlers.Signup(authSvc))
		r.Post("/auth/login", handlers.Login(authSvc))
		r.Post("/auth/logout", handlers.Logout(authSvc))
		r.With(middleware.AuthRequired(authSvc)).Get("/me", handlers.Me())
	})

	// NotFound handler returns JSON 404
	r.NotFound(handlers.HandleNotFound)

	return r
}
