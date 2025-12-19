package handlers

import (
	"net/http"

	"resume-tailor/internal/auth"
	"resume-tailor/internal/httpapi/cookies"
)

func Logout(authSvc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		token, ok := cookies.ReadSessionCookie(r)
		if ok == false {
			cookies.ClearSessionCookie(w)
			w.WriteHeader(http.StatusNoContent)
			return

		}

		err := authSvc.Logout(r.Context(), token)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return

		}

		cookies.ClearSessionCookie(w)
		w.WriteHeader(http.StatusNoContent)
	}
}
