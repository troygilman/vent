package route

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMountAndGroupStripPrefix(t *testing.T) {
	var gotPath string
	var gotID string

	root := New()
	if err := root.Mount("/admin/", func(admin *Router) {
		admin.Group("auth_users", func(users *Router) {
			users.GET("/{id}/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotID = r.PathValue("id")
				w.WriteHeader(http.StatusNoContent)
			}))
		})
		admin.Group("audit-events", func(events *Router) {
			events.GET("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.WriteHeader(http.StatusNoContent)
			}))
		})
	}); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/auth_users/42/", nil)
	root.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if gotPath != "/42/" {
		t.Fatalf("handler path = %q, want /42/", gotPath)
	}
	if gotID != "42" {
		t.Fatalf("PathValue(id) = %q, want 42", gotID)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin/audit-events/", nil)
	root.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("hyphenated group status = %d, want 204", rec.Code)
	}
	if gotPath != "/" {
		t.Fatalf("hyphenated group path = %q, want /", gotPath)
	}
}

func TestMiddlewareOnlyGroup(t *testing.T) {
	var order []string
	mw := func(label string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, label)
				next.ServeHTTP(w, r)
			})
		}
	}

	root := New()
	root.Use(mw("root"))
	if err := root.Mount("/admin/", func(admin *Router) {
		admin.Group("", func(authed *Router) {
			authed.Use(mw("authed"))
			authed.GET("/login/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "handler")
				w.WriteHeader(http.StatusNoContent)
			}))
		})
	}); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/login/", nil)
	root.Handler().ServeHTTP(rec, req)

	want := []string{"root", "authed", "handler"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v, want %v", order, want)
		}
	}
}

func TestRouteMiddlewareAppended(t *testing.T) {
	var order []string
	mw := func(label string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, label)
				next.ServeHTTP(w, r)
			})
		}
	}

	root := New()
	if err := root.Mount("/admin/", func(admin *Router) {
		admin.Use(mw("group"))
		admin.GET("/login/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "handler")
			w.WriteHeader(http.StatusNoContent)
		}), mw("route"))
	}); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/login/", nil)
	root.Handler().ServeHTTP(rec, req)

	want := []string{"group", "route", "handler"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v, want %v", order, want)
		}
	}
}

func TestNormalizePatternEnforcedOnGET(t *testing.T) {
	r := New()
	if err := r.GET("/Login/", http.NotFoundHandler()); err == nil {
		t.Fatal("expected error for invalid segment casing")
	}
}
