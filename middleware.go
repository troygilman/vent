package vent

import (
	"context"
	"net/http"
	"strconv"
	"vent/auth"
)

type claimsContextKey struct{}
type userContextKey struct{}

type Middleware = func(http.Handler) http.Handler

func NewAuthMiddleware(secretProvider auth.SecretProvider) Middleware {
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

			ctx := context.WithValue(r.Context(), claimsContextKey{}, claims)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func AuthorizationMiddleware(permissions []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(userContextKey{}).(EntityData)
			if !ok {
				http.Error(w, "user not found in context", http.StatusInternalServerError)
			}
			_ = user

			next.ServeHTTP(w, r)
		})
	}
}

func UserMiddleware(schema SchemaConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(claimsContextKey{}).(auth.VentClaims)
			if !ok {
				http.Error(w, "claims not present", http.StatusForbidden)
				return
			}

			userID, err := strconv.Atoi(claims.ID)
			if err != nil {
				http.Error(w, "claims ID is not an int", http.StatusInternalServerError)
			}

			entity, err := schema.Client.Get(r.Context(), userID)
			if err != nil {
				http.Error(w, "could not find user", http.StatusInternalServerError)
			}

			ctx := context.WithValue(r.Context(), userContextKey{}, entity)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
