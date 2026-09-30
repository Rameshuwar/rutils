package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"file-converter/internal/auth"
)

type contextKey string

const (
	// UserClaimsKey is the context key for authenticated JWT claims
	UserClaimsKey contextKey = "userClaims"
)

// RequireAuth middleware ensures the request has a valid Bearer JWT token
func RequireAuth(cfg *auth.Config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Authorization header is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Authorization header must be in format: Bearer <token>",
			})
			return
		}

		tokenStr := strings.TrimSpace(parts[1])
		claims, err := auth.ValidateToken(cfg, tokenStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid, expired, or malformed authentication token",
			})
			return
		}

		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		next(w, r.WithContext(ctx))
	}
}

// GetUserClaims retrieves the authenticated user's claims from the request context
func GetUserClaims(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*auth.Claims)
	return claims, ok
}
