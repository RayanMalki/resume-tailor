package cookies

import (
	"net/http"
	"time"
)

const SessionCookieName = "session"

func SetSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Expires:  expiresAt,
		HttpOnly: true,
		Path:     "/",
		Secure:   false, // for local dev set to false or browser will refuse the cookie
		SameSite: http.SameSiteLaxMode,
	})
}

func ReadSessionCookie(r *http.Request) (token string, ok bool) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return "", false
	}

	if cookie.Value == "" {
		return "", false
	}

	return cookie.Value, true
}

func ClearSessionCookie(w http.ResponseWriter) {
	c := &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false, // for local dev set to false or browser will refuse the cookie
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, c)
}


