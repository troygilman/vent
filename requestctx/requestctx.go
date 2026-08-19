package requestctx

import (
	"context"
	"net/http"

	"github.com/troygilman/vent/auth"
)

type adminPathKey struct{}
type csrfTokenKey struct{}
type themeKey struct{}

func WithAdminPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, adminPathKey{}, path)
}

func WithCSRFToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, csrfTokenKey{}, token)
}

func MustAdminPath(ctx context.Context) string {
	path, ok := ctx.Value(adminPathKey{}).(string)
	if !ok || path == "" {
		panic("requestctx: admin path not found in context")
	}
	return path
}

func MustCSRFToken(ctx context.Context) string {
	token, ok := ctx.Value(csrfTokenKey{}).(string)
	if !ok || token == "" {
		panic("requestctx: csrf token not found in context")
	}
	return token
}

func WithTheme(ctx context.Context, theme string) context.Context {
	return context.WithValue(ctx, themeKey{}, auth.NormalizeTheme(theme))
}

func MustTheme(ctx context.Context) string {
	theme, ok := ctx.Value(themeKey{}).(string)
	if !ok || theme == "" {
		panic("requestctx: theme not found in context")
	}
	return theme
}

// ThemeMiddleware reads the theme cookie into context and ensures a valid cookie
// exists on safe methods. The theme cookie is never cleared — only set or updated.
func ThemeMiddleware(secureCookies bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := themeFromCookie(r)
			theme := auth.NormalizeTheme(raw)

			if isSafeMethod(r.Method) {
				if raw == "" || raw != theme {
					auth.SetThemeCookie(w, theme, MustAdminPath(r.Context()), secureCookies)
				}
			}

			next.ServeHTTP(w, r.WithContext(WithTheme(r.Context(), theme)))
		})
	}
}

func themeFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(auth.ThemeCookieName)
	if err != nil || cookie.Value == "" {
		return ""
	}
	return cookie.Value
}

func AdminPathMiddleware(path string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := WithAdminPath(r.Context(), path)
			ctx = WithHTTPRequest(ctx, r)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type httpRequestKey struct{}

func WithHTTPRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, httpRequestKey{}, r)
}

func HTTPRequest(ctx context.Context) *http.Request {
	r, _ := ctx.Value(httpRequestKey{}).(*http.Request)
	return r
}

// RotateCSRFToken issues a new CSRF cookie. Call on auth-boundary events
// (successful login/logout) where a redirect follows so the next page embeds
// the new token. Do not rotate on ordinary CRUD mutations that re-render the
// same page — the response body would still carry the previous token.
func RotateCSRFToken(w http.ResponseWriter, r *http.Request, secureCookies bool) (string, error) {
	token, err := auth.NewCSRFToken()
	if err != nil {
		return "", err
	}
	auth.SetCSRFTokenCookie(w, token, MustAdminPath(r.Context()), secureCookies)
	return token, nil
}

// CSRFMiddleware manages CSRF tokens for admin requests.
// Safe methods issue a token when missing and reuse an existing cookie.
// Mutating methods require a valid cookie and matching X-CSRF-Token header,
// then load the token into context for templates and error re-renders.
func CSRFMiddleware(secureCookies bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isSafeMethod(r.Method) {
				token := csrfTokenFromCookie(r)
				if token == "" {
					var err error
					token, err = auth.NewCSRFToken()
					if err != nil {
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}
					auth.SetCSRFTokenCookie(w, token, MustAdminPath(r.Context()), secureCookies)
				}
				next.ServeHTTP(w, r.WithContext(WithCSRFToken(r.Context(), token)))
				return
			}

			if !auth.ValidateCSRFToken(r) {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			token := csrfTokenFromCookie(r)
			next.ServeHTTP(w, r.WithContext(WithCSRFToken(r.Context(), token)))
		})
	}
}

func csrfTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(auth.CSRFTokenCookieName)
	if err != nil || cookie.Value == "" {
		return ""
	}
	return cookie.Value
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}
