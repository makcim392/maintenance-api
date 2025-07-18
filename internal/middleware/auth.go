package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/makcim392/maintenance-api/internal/auth"
)

const (
	userIDContextKey contextKey = "userID"
	roleContextKey   contextKey = "role"
)

type AuthMiddlewareHandler struct {
	validator auth.TokenValidator
}

func NewAuthMiddlewareHandler(validator auth.TokenValidator) *AuthMiddlewareHandler {
	return &AuthMiddlewareHandler{
		validator: validator,
	}
}

func (h *AuthMiddlewareHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		claims, err := h.validator.ValidateToken(bearerToken[1])
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, int(claims.UserID))
		ctx = context.WithValue(ctx, roleContextKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
