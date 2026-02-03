package cookies

import (
	"net/http"
	"os"
	"strings"
	"time"
)

const SessionCookieName = "session"
const OAuthStateCookieName = "oauth_state"
const OAuthRedirectCookieName = "oauth_redirect"

func SetSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	secure, sameSite := sessionCookieOptions()
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Expires:  expiresAt,
		HttpOnly: true,
		Path:     "/",
		Secure:   secure,
		SameSite: sameSite,
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
	secure, sameSite := sessionCookieOptions()
	c := &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	}

	http.SetCookie(w, c)
}

func SetOAuthStateCookie(w http.ResponseWriter, state string, expiresAt time.Time) {
	secure, sameSite := sessionCookieOptions()
	http.SetCookie(w, &http.Cookie{
		Name:     OAuthStateCookieName,
		Value:    state,
		Expires:  expiresAt,
		HttpOnly: true,
		Path:     "/",
		Secure:   secure,
		SameSite: sameSite,
	})
}

func ReadOAuthStateCookie(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(OAuthStateCookieName)
	if err != nil {
		return "", false
	}
	if cookie.Value == "" {
		return "", false
	}
	return cookie.Value, true
}

func ClearOAuthStateCookie(w http.ResponseWriter) {
	secure, sameSite := sessionCookieOptions()
	c := &http.Cookie{
		Name:     OAuthStateCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	}
	http.SetCookie(w, c)
}

func SetOAuthRedirectCookie(w http.ResponseWriter, path string, expiresAt time.Time) {
	secure, sameSite := sessionCookieOptions()
	http.SetCookie(w, &http.Cookie{
		Name:     OAuthRedirectCookieName,
		Value:    path,
		Expires:  expiresAt,
		HttpOnly: true,
		Path:     "/",
		Secure:   secure,
		SameSite: sameSite,
	})
}

func ReadOAuthRedirectCookie(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(OAuthRedirectCookieName)
	if err != nil {
		return "", false
	}
	if cookie.Value == "" {
		return "", false
	}
	return cookie.Value, true
}

func ClearOAuthRedirectCookie(w http.ResponseWriter) {
	secure, sameSite := sessionCookieOptions()
	c := &http.Cookie{
		Name:     OAuthRedirectCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	}
	http.SetCookie(w, c)
}

func sessionCookieOptions() (bool, http.SameSite) {
	secure := envTruthy("COOKIE_SECURE")
	sameSite := parseSameSite(os.Getenv("COOKIE_SAMESITE"))
	if sameSite == http.SameSiteNoneMode && !secure {
		secure = true
	}
	return secure, sameSite
}

func envTruthy(key string) bool {
	val := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	return val == "1" || val == "true" || val == "yes"
}

func parseSameSite(value string) http.SameSite {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	case "lax":
		fallthrough
	default:
		return http.SameSiteLaxMode
	}
}
