package route

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeSegment(t *testing.T) {
	t.Run("empty allowed", func(t *testing.T) {
		got, err := NormalizeSegment("")
		if err != nil || got != "" {
			t.Fatalf("NormalizeSegment(\"\") = (%q, %v)", got, err)
		}
	})

	t.Run("valid", func(t *testing.T) {
		got, err := NormalizeSegment("auth_users")
		if err != nil || got != "auth_users" {
			t.Fatalf("NormalizeSegment() = (%q, %v)", got, err)
		}
	})

	t.Run("rejects slashes", func(t *testing.T) {
		if _, err := NormalizeSegment("/auth_users"); err == nil {
			t.Fatal("expected error for leading slash")
		}
	})

	t.Run("accepts hyphens", func(t *testing.T) {
		got, err := NormalizeSegment("audit-events")
		if err != nil || got != "audit-events" {
			t.Fatalf("NormalizeSegment(audit-events) = (%q, %v)", got, err)
		}
	})

	t.Run("rejects uppercase", func(t *testing.T) {
		if _, err := NormalizeSegment("AuthUser"); err == nil {
			t.Fatal("expected error for uppercase")
		}
	})
}

func TestNormalizePattern(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "/"},
		{"/", "/"},
		{"/login", "/login/"},
		{"/{id}", "/{id}/"},
		{"/{id}/{$}", "/{id}/{$}"},
		{"/add/{$}", "/add/{$}"},
		{"/{id}/password", "/{id}/password/"},
		{"/options/groups", "/options/groups/"},
		{"/options/user", "/options/user/"},
	}

	for _, tt := range tests {
		got, err := NormalizePattern(tt.in)
		if err != nil {
			t.Fatalf("NormalizePattern(%q) error: %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("NormalizePattern(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestNormalizePatternRejectsUnknownParam(t *testing.T) {
	if _, err := NormalizePattern("/{slug}/"); err == nil {
		t.Fatal("expected error for unknown parameter")
	}
	if _, err := NormalizePattern("/options/{edge}/"); err == nil {
		t.Fatal("expected error for {edge} parameter")
	}
}

func TestPasswordAndWildcardOptionsConflict(t *testing.T) {
	root := New()
	if err := root.Group("users", func(users *Router) {
		_ = users.GET("/{id}/password/", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		if err := users.GET("/options/{edge}/", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})); err == nil {
			t.Fatal("expected NormalizePattern error for {edge}")
		}
	}); err != nil {
		t.Fatalf("group: %v", err)
	}
}

func TestPasswordAndLiteralOptionsRegister(t *testing.T) {
	root := New()
	if err := root.Group("users", func(users *Router) {
		users.GET("/{id}/password/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		users.GET("/options/groups/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		}))
	}); err != nil {
		t.Fatalf("literal options route with password routes: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/options/groups/", nil)
	root.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("options status = %d, want 202", rec.Code)
	}
}

func TestLiteralOptionsDoNotShareAHandler(t *testing.T) {
	root := New()
	if err := root.Group("reviews", func(schema *Router) {
		schema.GET("/options/user/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("user"))
		}))
		schema.GET("/options/book/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("book"))
		}))
		schema.GET("/{id}/{$}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/reviews/options/user/", nil)
	root.Handler().ServeHTTP(rec, req)
	if rec.Body.String() != "user" {
		t.Fatalf("first edge = %q, want user", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/reviews/options/book/", nil)
	root.Handler().ServeHTTP(rec, req)
	if rec.Body.String() != "book" {
		t.Fatalf("second edge = %q, want book", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/reviews/options/not-an-edge/", nil)
	root.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown edge status = %d, want 404", rec.Code)
	}
}

func TestNormalizeMountPrefix(t *testing.T) {
	mount, strip, err := NormalizeMountPrefix("/admin")
	if err != nil {
		t.Fatalf("NormalizeMountPrefix error: %v", err)
	}
	if mount != "/admin/" || strip != "/admin" {
		t.Fatalf("mount=%q strip=%q", mount, strip)
	}
}
