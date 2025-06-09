package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"felton.com/microservicecomm/auth"
	response "felton.com/microservicecomm/transport"
)

type contextKey string

const UserContextKey = contextKey("user_claims")

func Auth(validator *auth.TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.ResponseWithError(w, http.StatusUnauthorized, "Authorization header required")
				return
			}
			parts := strings.Split(authHeader, " ")
			slog.Info("parts", "parts", parts)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				response.ResponseWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}
			tokenString := parts[1]

			slog.Info("tokenString", "tokenString", tokenString)

			claims, err := validator.Validate(tokenString)
			if err != nil {
				response.ResponseWithError(w, http.StatusUnauthorized, "Invalid token: "+err.Error())
				return
			}
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper to get claims from context
func GetClaims(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*auth.Claims)
	return claims, ok
}
