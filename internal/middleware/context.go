package middleware

// Define custom types for context keys
type contextKey string

const (
	UserIDContextKey contextKey = "userID"
	RoleContextKey   contextKey = "role"
)
