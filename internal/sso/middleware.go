package sso

import (
	"context"
	"net/http"

	"github.com/shrihariharanba/go-gateway/internal/sso/providers"
)

type contextKey string

const AuthContextKey contextKey = "auth"

func AuthMiddleware(provider providers.SSOProvider, authRequired bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var authCtx *providers.AuthContext
			if provider != nil {
				token := r.Header.Get("Authorization")
				if token == "" && authRequired {
					http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
					return
				}
				var err error
				authCtx, err = provider.Authenticate(r.Context(), token)
				if err != nil && authRequired {
					http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
					return
				}
			}

			if authCtx == nil {
				authCtx = &providers.AuthContext{
					UserID: "anonymous",
					Roles:  []string{"public"},
				}
			}

			ctx := context.WithValue(r.Context(), AuthContextKey, authCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func FromContext(ctx context.Context) *providers.AuthContext {
	if v := ctx.Value(AuthContextKey); v != nil {
		if auth, ok := v.(*providers.AuthContext); ok {
			return auth
		}
	}
	return &providers.AuthContext{
		UserID: "anonymous",
		Roles:  []string{"public"},
	}
}
