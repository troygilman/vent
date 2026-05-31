package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeTheme(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{ThemeLight, ThemeLight},
		{ThemeDark, ThemeDark},
		{ThemeSystem, ThemeSystem},
		{"", ThemeSystem},
		{"invalid", ThemeSystem},
	}

	for _, tt := range tests {
		if got := NormalizeTheme(tt.in); got != tt.want {
			t.Fatalf("NormalizeTheme(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestNextTheme(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{ThemeSystem, ThemeLight},
		{ThemeLight, ThemeDark},
		{ThemeDark, ThemeSystem},
		{"invalid", ThemeLight},
	}

	for _, tt := range tests {
		if got := NextTheme(tt.in); got != tt.want {
			t.Fatalf("NextTheme(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestThemeLabel(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{ThemeSystem, "System"},
		{ThemeLight, "Light"},
		{ThemeDark, "Dark"},
		{"invalid", "System"},
	}

	for _, tt := range tests {
		if got := ThemeLabel(tt.in); got != tt.want {
			t.Fatalf("ThemeLabel(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSetThemeCookie(t *testing.T) {
	rec := httptest.NewRecorder()
	SetThemeCookie(rec, "dark", "/admin/", true)

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookie count = %d, want 1", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != ThemeCookieName {
		t.Fatalf("cookie name = %q, want %q", cookie.Name, ThemeCookieName)
	}
	if cookie.Value != ThemeDark {
		t.Fatalf("cookie value = %q, want %q", cookie.Value, ThemeDark)
	}
	if cookie.Path != "/admin/" {
		t.Fatalf("cookie path = %q, want %q", cookie.Path, "/admin/")
	}
	if !cookie.HttpOnly {
		t.Fatal("theme cookie should be HttpOnly")
	}
	if !cookie.Secure {
		t.Fatal("expected Secure flag when secure=true")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("SameSite = %v, want Lax", cookie.SameSite)
	}
	if cookie.MaxAge != ThemeCookieMaxAge {
		t.Fatalf("MaxAge = %d, want %d", cookie.MaxAge, ThemeCookieMaxAge)
	}
}

func TestSetThemeCookieNormalizesInvalidValue(t *testing.T) {
	rec := httptest.NewRecorder()
	SetThemeCookie(rec, "bogus", "/admin/", false)

	cookie := rec.Result().Cookies()[0]
	if cookie.Value != ThemeSystem {
		t.Fatalf("cookie value = %q, want %q", cookie.Value, ThemeSystem)
	}
}
