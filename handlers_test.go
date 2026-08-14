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
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("Cache-Control = %q, want immutable long cache", got)
	}
}

func TestStaticDirHandlerServesFont(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/fonts/inter-latin.woff2", nil)
	StaticDirHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=3600" {
		t.Fatalf("Cache-Control = %q, want public, max-age=3600", got)
	}
}
