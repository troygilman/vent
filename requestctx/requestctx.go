package requestctx

import (
	"context"
	"net/http"

	"github.com/troygilman/vent/auth"
)

type adminPathKey struct{}
type csrfTokenKey struct{}

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

func AdminPathMiddleware(path string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := WithAdminPath(r.Context(), path)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CSRFTokenMiddleware(secureCookies bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := auth.NewCSRFToken()
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			auth.SetCSRFTokenCookie(w, token, MustAdminPath(r.Context()), secureCookies)
			ctx := WithCSRFToken(r.Context(), token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CSRFLoadMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(auth.CSRFTokenCookieName)
			if err == nil && cookie.Value != "" {
				r = r.WithContext(WithCSRFToken(r.Context(), cookie.Value))
			}
			next.ServeHTTP(w, r)
		})
	}
}

func CSRFValidationMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !auth.ValidateCSRFToken(r) {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
