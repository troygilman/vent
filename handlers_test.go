package vent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStaticAssetVersion(t *testing.T) {
	if StaticAssetVersion == "" {
		t.Fatal("StaticAssetVersion is empty")
	}
	if len(StaticAssetVersion) != 16 {
		t.Fatalf("StaticAssetVersion = %q, want 16 hex chars", StaticAssetVersion)
	}
}

func TestStaticDirHandlerServesJS(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/js/datastar-plugins.js", nil)
	StaticDirHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `name: "confirm"`) {
		t.Fatal("datastar-plugins.js does not contain confirm plugin")
	}
	if !strings.Contains(rec.Body.String(), `name: "cookie"`) {
		t.Fatal("datastar-plugins.js does not contain cookie persist plugin")
	}
	if !strings.Contains(rec.Body.String(), `key: "must"`) {
		t.Fatal("cookie plugin should require a cookie name key")
	}
	if strings.Contains(rec.Body.String(), "`vent-${key}`") || strings.Contains(rec.Body.String(), "vent-${key}") {
		t.Fatal("cookie plugin should not prefix cookie names with vent-")
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("Cache-Control = %q, want immutable long cache", got)
	}
}

func TestStaticDirHandlerCacheControl(t *testing.T) {
	paths := []string{
		"/static/js/common.js",
		"/static/css/style.css",
		"/static/fonts/inter-latin.woff2",
		"/static/img/favicon.ico",
		"/static/img/vent-logo.png",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			StaticDirHandler().ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
				t.Fatalf("Cache-Control = %q, want immutable long cache", got)
			}
		})
	}
}
