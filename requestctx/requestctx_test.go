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

func TestCSRFMiddlewareIssuesTokenWhenNoCookie(t *testing.T) {
	var got string
	handler := AdminPathMiddleware("/admin/")(
		CSRFMiddleware(false)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = MustCSRFToken(r.Context())
				w.WriteHeader(http.StatusNoContent)
			}),
		),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/login/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got == "" {
		t.Fatal("expected csrf token in context")
	}
	if len(got) != 64 {
		t.Fatalf("token length = %d, want 64", len(got))
	}
	if findCSRFCookie(rec) == nil || findCSRFCookie(rec).Value != got {
		t.Fatal("expected csrf cookie to match context token")
	}
}

func TestCSRFMiddlewareRotatesTokenOnGET(t *testing.T) {
	handler := AdminPathMiddleware("/admin/")(
		CSRFMiddleware(false)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if MustCSRFToken(r.Context()) == "existing-token" {
					t.Fatal("GET should rotate csrf token, not reuse cookie value")
				}
				w.WriteHeader(http.StatusNoContent)
			}),
		),
	)

	req := httptest.NewRequest(http.MethodGet, "/admin/login/", nil)
	req.AddCookie(&http.Cookie{Name: auth.CSRFTokenCookieName, Value: "existing-token"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	cookie := findCSRFCookie(rec)
	if cookie == nil || cookie.Value == "existing-token" {
		t.Fatal("expected rotated csrf cookie on GET")
	}
}

func TestCSRFMiddlewareLoadsExistingCookieOnPOST(t *testing.T) {
	handler := AdminPathMiddleware("/admin/")(
		CSRFMiddleware(false)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if MustCSRFToken(r.Context()) != "existing-token" {
					t.Fatalf("unexpected csrf token in context")
				}
				w.WriteHeader(http.StatusNoContent)
			}),
		),
	)

	t.Run(http.MethodPost, func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/admin/login/", nil)
		req.AddCookie(&http.Cookie{Name: auth.CSRFTokenCookieName, Value: "existing-token"})
		req.Header.Set(auth.CSRFTokenHeaderName, "existing-token")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	})
}

func TestCSRFMiddlewareRejectsPOSTWithoutValidToken(t *testing.T) {
	called := false
	handler := AdminPathMiddleware("/admin/")(
		CSRFMiddleware(false)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusNoContent)
			}),
		),
	)

	req := httptest.NewRequest(http.MethodPost, "/admin/login/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if called {
		t.Fatal("handler was called without valid csrf token")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestCSRFMiddlewareRejectsPOSTWithMismatchedToken(t *testing.T) {
	called := false
	handler := AdminPathMiddleware("/admin/")(
		CSRFMiddleware(false)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusNoContent)
			}),
		),
	)

	req := httptest.NewRequest(http.MethodPost, "/admin/login/", nil)
	req.AddCookie(&http.Cookie{Name: auth.CSRFTokenCookieName, Value: "cookie-token"})
	req.Header.Set(auth.CSRFTokenHeaderName, "header-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if called {
		t.Fatal("handler was called with mismatched csrf token")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func findCSRFCookie(rec *httptest.ResponseRecorder) *http.Cookie {
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == auth.CSRFTokenCookieName {
			return cookie
		}
	}
	return nil
}

func TestMustTheme(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		ctx := WithTheme(context.Background(), auth.ThemeDark)
		if got := MustTheme(ctx); got != auth.ThemeDark {
			t.Fatalf("MustTheme() = %q, want %q", got, auth.ThemeDark)
		}
	})

	t.Run("panics when missing", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		MustTheme(context.Background())
	})
}

func TestThemeMiddlewareSetsDefaultCookie(t *testing.T) {
	var got string
	handler := AdminPathMiddleware("/admin/")(
		ThemeMiddleware(false)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = MustTheme(r.Context())
				w.WriteHeader(http.StatusNoContent)
			}),
		),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	handler.ServeHTTP(rec, req)

	if got != auth.ThemeSystem {
		t.Fatalf("theme = %q, want %q", got, auth.ThemeSystem)
	}
	cookie := findThemeCookie(rec)
	if cookie == nil || cookie.Value != auth.ThemeSystem {
		t.Fatal("expected default theme cookie to be set")
	}
}

func TestThemeMiddlewareReadsExistingCookie(t *testing.T) {
	var got string
	handler := AdminPathMiddleware("/admin/")(
		ThemeMiddleware(false)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = MustTheme(r.Context())
				w.WriteHeader(http.StatusNoContent)
			}),
		),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	req.AddCookie(&http.Cookie{Name: auth.ThemeCookieName, Value: auth.ThemeDark})
	handler.ServeHTTP(rec, req)

	if got != auth.ThemeDark {
		t.Fatalf("theme = %q, want %q", got, auth.ThemeDark)
	}
	if findThemeCookie(rec) != nil {
		t.Fatal("expected valid theme cookie to be left unchanged on GET")
	}
}

func TestThemeMiddlewareFixesInvalidCookieOnGET(t *testing.T) {
	handler := AdminPathMiddleware("/admin/")(
		ThemeMiddleware(false)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if MustTheme(r.Context()) != auth.ThemeSystem {
					t.Fatalf("unexpected theme in context")
				}
				w.WriteHeader(http.StatusNoContent)
			}),
		),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	req.AddCookie(&http.Cookie{Name: auth.ThemeCookieName, Value: "neon"})
	handler.ServeHTTP(rec, req)

	cookie := findThemeCookie(rec)
	if cookie == nil || cookie.Value != auth.ThemeSystem {
		t.Fatal("expected invalid theme cookie to be normalized on GET")
	}
}

func TestThemeMiddlewareDoesNotSetCookieOnPOST(t *testing.T) {
	handler := AdminPathMiddleware("/admin/")(
		ThemeMiddleware(false)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if MustTheme(r.Context()) != auth.ThemeSystem {
					t.Fatalf("unexpected theme in context")
				}
				w.WriteHeader(http.StatusNoContent)
			}),
		),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/preferences/theme/", nil)
	handler.ServeHTTP(rec, req)

	if findThemeCookie(rec) != nil {
		t.Fatal("expected theme middleware not to set cookie on POST")
	}
}

func findThemeCookie(rec *httptest.ResponseRecorder) *http.Cookie {
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == auth.ThemeCookieName {
			return cookie
		}
	}
	return nil
}
