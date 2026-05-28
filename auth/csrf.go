package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
)

const (
	CSRFTokenCookieName = "vent-csrf-token"
	CSRFTokenHeaderName = "X-CSRF-Token"
	// CSRFTokenMaxAge matches the auth session lifetime.
	CSRFTokenMaxAge = 24 * 60 * 60
)

func NewCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func SetCSRFTokenCookie(w http.ResponseWriter, token, path string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFTokenCookieName,
		Value:    token,
		Path:     path,
		MaxAge:   CSRFTokenMaxAge,
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func ValidateCSRFToken(r *http.Request) bool {
	cookie, err := r.Cookie(CSRFTokenCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}

	headerToken := r.Header.Get(CSRFTokenHeaderName)
	if headerToken == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(headerToken)) == 1
}
