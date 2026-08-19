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
	if got := rec.Header().Get("Cache-Control"); got != immutableCacheControl {
		t.Fatalf("Cache-Control = %q, want immutable long cache", got)
	}
}

func TestStaticDirHandlerCacheControl(t *testing.T) {
	tests := []struct {
		path  string
		code  int
		cache string
	}{
		{path: "/static/js/common.js", code: http.StatusOK, cache: immutableCacheControl},
		{path: "/static/css/style.css", code: http.StatusOK, cache: immutableCacheControl},
		{path: "/static/fonts/inter-latin.woff2", code: http.StatusOK, cache: immutableCacheControl},
		{path: "/static/img/favicon.ico", code: http.StatusOK, cache: immutableCacheControl},
		{path: "/static/img/vent-logo.png", code: http.StatusOK, cache: immutableCacheControl},
		{path: "/static/missing.bin", code: http.StatusNotFound, cache: "no-store"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			StaticDirHandler().ServeHTTP(rec, req)
			if rec.Code != tt.code {
				t.Fatalf("status = %d, want %d", rec.Code, tt.code)
			}
			if got := rec.Header().Get("Cache-Control"); got != tt.cache {
				t.Fatalf("Cache-Control = %q, want %q", got, tt.cache)
			}
		})
	}
}

func TestStaticDirHandlerVersionsCSSURLs(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/css/style.css", nil)
	StaticDirHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	wantFont := `url("../fonts/inter-latin.woff2?v=` + StaticAssetVersion + `")`
	wantLogo := `url("../img/vent-logo.png?v=` + StaticAssetVersion + `")`
	if !strings.Contains(body, wantFont) {
		t.Fatalf("css missing versioned font url %q", wantFont)
	}
	if !strings.Contains(body, wantLogo) {
		t.Fatalf("css missing versioned logo url %q", wantLogo)
	}
	if strings.Contains(body, `url("../fonts/inter-latin.woff2")`) {
		t.Fatal("css still has unversioned font url")
	}
	if strings.Contains(body, `url("../img/vent-logo.png")`) {
		t.Fatal("css still has unversioned logo url")
	}
}
