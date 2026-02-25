package vent

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"vent/auth"
)

type claimsContextKey struct{}
type userContextKey struct{}

type Middleware = func(http.Handler) http.Handler

func AuthentificationMiddleware(authenticator auth.TokenAuthenticator) Middleware {
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

func UserMiddleware(schema SchemaConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(claimsContextKey{}).(*auth.VentClaims)
			if !ok {
				http.Error(w, "claims not present", http.StatusForbidden)
				return
			}

			userID, err := strconv.Atoi(claims.Subject)
			if err != nil {
				http.Error(w, "claims ID is not an int", http.StatusInternalServerError)
				return
			}

			entity, err := schema.Client.Get(r.Context(), userID, GetOptions{
				WithEdges: []string{
					"groups__permissions",
				},
			})
			if err != nil {
				http.Error(w, "could not find user", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey{}, entity)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func AuthorizationMiddleware(permissions ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(userContextKey{}).(EntityData)
			if !ok {
				http.Error(w, "user not found in context", http.StatusInternalServerError)
				return
			}

			if !user.GetBool("is_superuser") {
				userPermissions := make(map[string]struct{})
				groups := user.GetEdges("groups")
				for _, group := range groups {
					for _, permission := range group.GetEdges("permissions") {
						userPermissions[permission.GetString("name")] = struct{}{}
					}
				}

				for _, permission := range permissions {
					if _, ok := userPermissions[permission]; !ok {
						http.Error(w, fmt.Sprintf("user does not have permission: %s", permission), http.StatusUnauthorized)
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func LoggerMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s\n", r.Method, r.RequestURI)
			next.ServeHTTP(w, r)
		})
	}
}
