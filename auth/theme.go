package auth

import "net/http"

const (
	ThemeCookieName = "vent-theme"
	ThemeSystem     = "system"
	ThemeLight      = "light"
	ThemeDark       = "dark"
	// ThemeCookieMaxAge is one year — a long-lived UI preference, not a session.
	ThemeCookieMaxAge = 365 * 24 * 60 * 60
)

func NormalizeTheme(value string) string {
	switch value {
	case ThemeLight, ThemeDark, ThemeSystem:
		return value
	default:
		return ThemeSystem
	}
}

func NextTheme(current string) string {
	switch NormalizeTheme(current) {
	case ThemeSystem:
		return ThemeLight
	case ThemeLight:
		return ThemeDark
	default:
		return ThemeSystem
	}
}

func ThemeLabel(theme string) string {
	switch NormalizeTheme(theme) {
	case ThemeLight:
		return "Light"
	case ThemeDark:
		return "Dark"
	default:
		return "System"
	}
}

func SetThemeCookie(w http.ResponseWriter, theme, path string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     ThemeCookieName,
		Value:    NormalizeTheme(theme),
		Path:     path,
		MaxAge:   ThemeCookieMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}
