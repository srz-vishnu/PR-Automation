package middleware

import (
	"context"
	"net/http"
	"pr-mail/pkg/api"
	"pr-mail/pkg/jwt"
	"strings"
)

type contextKey string

const (
	AdminIDKey   contextKey = "admin_id"
	AdminNameKey contextKey = "admin_username"
)

// JWTAuthMiddleware for verifying admin JWT token
func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			api.Fail(w, http.StatusUnauthorized, 401, "Authorization header is missing", "")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			api.Fail(w, http.StatusUnauthorized, 401, "Invalid token format", "")
			return
		}

		claims, err := jwt.ValidateToken(tokenString)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				api.Fail(w, http.StatusUnauthorized, 401, "Token expired, please log in again", "")
				return
			}
			api.Fail(w, http.StatusUnauthorized, 401, "Invalid token", err.Error())
			return
		}

		// Store admin ID and name in context
		ctx := context.WithValue(r.Context(), AdminIDKey, claims.AdminID)
		ctx = context.WithValue(ctx, AdminNameKey, claims.Username)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
