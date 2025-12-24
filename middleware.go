package vent

import (
	"context"
	"net/http"
	"vent/auth"
)

type Middleware = func(http.Handler) http.Handler

func AuthMiddleware(secretProvider auth.SecretProvider) Middleware {
	authenticator := auth.NewJwtTokenAuthenticator(secretProvider)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenCookie, err := r.Cookie("vent-auth-token")
			if err != nil {
				http.Redirect(w, r, "/admin/login/", http.StatusSeeOther)
				return
			}

			claims, err := authenticator.Authenticate(tokenCookie.Value)
			if err != nil {
				http.Redirect(w, r, "/admin/login/", http.StatusSeeOther)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, "claims", claims)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
