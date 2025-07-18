package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/makcim392/maintenance-api/internal/auth"
	"github.com/stretchr/testify/assert"
)

// MockTokenValidator implements auth.TokenValidator for testing
type MockTokenValidator struct {
	validateFunc func(string) (*auth.Claims, error)
}

func (m *MockTokenValidator) ValidateToken(tokenString string) (*auth.Claims, error) {
	if m.validateFunc != nil {
		return m.validateFunc(tokenString)
	}
	return nil, fmt.Errorf("mock not implemented")
}

func TestAuthMiddleware(t *testing.T) {
	// Mock handler to use with the middleware
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(userIDContextKey).(int)
		role := r.Context().Value(roleContextKey).(string)

		assert.Greater(t, userID, 0)
		assert.NotEmpty(t, role)

		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name         string
		authHeader   string
		expectedCode int
		validator    auth.TokenValidator
	}{
		{
			name:         "Valid token",
			authHeader:   "Bearer valid-token",
			expectedCode: http.StatusOK,
			validator: &MockTokenValidator{
				validateFunc: func(token string) (*auth.Claims, error) {
					return &auth.Claims{
						UserID: 123,
						Role:   "user",
					}, nil
				},
			},
		},
		{
			name:         "Missing authorization header",
			authHeader:   "",
			expectedCode: http.StatusUnauthorized,
			validator:    &MockTokenValidator{},
		},
		{
			name:         "Invalid token format",
			authHeader:   "InvalidFormat",
			expectedCode: http.StatusUnauthorized,
			validator:    &MockTokenValidator{},
		},
		{
			name:         "Invalid token",
			authHeader:   "Bearer invalid-token",
			expectedCode: http.StatusUnauthorized,
			validator: &MockTokenValidator{
				validateFunc: func(token string) (*auth.Claims, error) {
					return nil, fmt.Errorf("invalid token")
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware handler with the mock validator
			middlewareHandler := NewAuthMiddlewareHandler(tt.validator)

			// Create a new request with the test case's authorization header
			req := httptest.NewRequest("GET", "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Create and execute the middleware
			handler := middlewareHandler.AuthMiddleware(nextHandler)
			handler.ServeHTTP(rr, req)

			// Assert the response status code matches expected
			assert.Equal(t, tt.expectedCode, rr.Code)
		})
	}
}
