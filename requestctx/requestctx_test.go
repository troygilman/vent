package requestctx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/troygilman/vent/auth"
)

func TestMustAdminPath(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		ctx := WithAdminPath(context.Background(), "/admin/")
		if got := MustAdminPath(ctx); got != "/admin/" {
			t.Fatalf("MustAdminPath() = %q, want /admin/", got)
		}
	})

	t.Run("panics when missing", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		MustAdminPath(context.Background())
	})

	t.Run("panics when empty", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		MustAdminPath(WithAdminPath(context.Background(), ""))
	})
}

func TestMustCSRFToken(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		ctx := WithCSRFToken(context.Background(), "token")
		if got := MustCSRFToken(ctx); got != "token" {
			t.Fatalf("MustCSRFToken() = %q, want token", got)
		}
	})

	t.Run("panics when missing", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		MustCSRFToken(context.Background())
	})
}

func TestAdminPathMiddleware(t *testing.T) {
	var got string
	handler := AdminPathMiddleware("/admin/")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = MustAdminPath(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got != "/admin/" {
		t.Fatalf("admin path = %q, want /admin/", got)
	}
}

func TestCSRFTokenMiddleware(t *testing.T) {
	var got string
	handler := AdminPathMiddleware("/admin/")(
		CSRFTokenMiddleware(false)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = MustCSRFToken(r.Context())
			}),
		),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/login/", nil)
	handler.ServeHTTP(rec, req)

	if got == "" {
		t.Fatal("expected csrf token in context")
	}
	cookies := rec.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == auth.CSRFTokenCookieName && c.Value == got {
			found = true
		}
	}
	if !found {
		t.Fatal("expected csrf cookie to match context token")
	}
}

func TestCSRFLoadMiddleware(t *testing.T) {
	handler := AdminPathMiddleware("/admin/")(
		CSRFLoadMiddleware()(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if MustCSRFToken(r.Context()) != "existing-token" {
					t.Fatalf("unexpected csrf token in context")
				}
			}),
		),
	)

	req := httptest.NewRequest(http.MethodPost, "/admin/login/", nil)
	req.AddCookie(&http.Cookie{Name: auth.CSRFTokenCookieName, Value: "existing-token"})
	handler.ServeHTTP(httptest.NewRecorder(), req)
}
