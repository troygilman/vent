package route

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouteRegisteredBeforeUseSkipsLaterMiddleware(t *testing.T) {
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
		admin.Use(mw("base"))
		admin.Handle("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		admin.Use(mw("csrf"))
		admin.GET("/login/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
	}); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	rec := httptest.NewRecorder()
	root.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/static/app.js", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("static status = %d, want 204", rec.Code)
	}
	if len(order) != 1 || order[0] != "base" {
		t.Fatalf("static middleware order = %v, want [base]", order)
	}

	order = nil
	rec = httptest.NewRecorder()
	root.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/login/", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("login status = %d, want 204", rec.Code)
	}
	if len(order) != 2 || order[0] != "base" || order[1] != "csrf" {
		t.Fatalf("login middleware order = %v, want [base csrf]", order)
	}
}
