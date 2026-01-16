package auth

import (
	"context"
	"net/http"
	"strings"
)

// AuthMiddleware returns a middleware enforcing JWT validation
func AuthMiddleware(oidcProvider *OIDCProvider, required bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If auth is optional and no token, just skip
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				if !required {
					next.ServeHTTP(w, r)
					return
				}
				http.Error(w, "Authorization header missing", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
				return
			}
			rawToken := parts[1]

			// Validate token
			idToken, err := oidcProvider.ValidateToken(r.Context(), rawToken)
			if err != nil {
				http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}

			// Extract claims and store in context
			claims, err := ExtractClaims(idToken)
			if err != nil {
				http.Error(w, "Failed to parse token claims", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ClaimsKey, claims)

			// Continue to next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper to retrieve claims from context
func GetClaims(r *http.Request) map[string]interface{} {
	if c, ok := r.Context().Value(ClaimsKey).(map[string]interface{}); ok {
		return c
	}
	return nil
}
