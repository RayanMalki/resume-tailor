package httpapi

import (
	"net/http"

	"resume-tailor/internal/auth"
	"resume-tailor/internal/httpapi/handlers"
	"resume-tailor/internal/httpapi/middleware"
	"resume-tailor/internal/resumes"
	"resume-tailor/internal/runs"

	"github.com/go-chi/chi/v5"
)

func NewRouter(authSvc *auth.Service, runsSvc *runs.Service, resumesSvc *resumes.Service) http.Handler {
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

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthRequired(authSvc))

			//GET request
			r.Get("/me", handlers.Me())
			r.Get("/runs/{runID}", handlers.GetRunByIdHandler(runsSvc))
			r.Get("/runs", handlers.ListRunsHandler(runsSvc))

			//POST request
			r.Post("/runs", handlers.CreateRunHandler(runsSvc, resumesSvc))
			r.Post("/resumes", handlers.CreateResumeHandler(resumesSvc))
		})

	})

	// NotFound handler returns JSON 404
	r.NotFound(handlers.HandleNotFound)

	return r
}
