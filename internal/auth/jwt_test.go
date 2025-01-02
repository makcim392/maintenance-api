package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

// Existing TestGenerateToken and TestValidateToken remain the same...

func TestJWTValidator_ValidateToken(t *testing.T) {
	validator := &JWTValidator{}

	tests := []struct {
		name       string
		setupToken func() string
		wantUserID uint
		wantRole   string
		wantErr    bool
	}{
		{
			name: "Valid token with JWTValidator",
			setupToken: func() string {
				token, _ := GenerateToken(1, "user")
				return token
			},
			wantUserID: 1,
			wantRole:   "user",
			wantErr:    false,
		},
		{
			name: "Invalid signing method",
			setupToken: func() string {
				claims := Claims{
					UserID: 1,
					Role:   "user",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				// Sign with HS256 key but claim RS256
				tokenString, _ := token.SignedString(jwtKey)
				return tokenString
			},
			wantUserID: 0,
			wantRole:   "",
			wantErr:    true,
		},
		{
			name: "Malformed token",
			setupToken: func() string {
				return "header.payload" // Missing signature part
			},
			wantUserID: 0,
			wantRole:   "",
			wantErr:    true,
		},
		{
			name: "Token with invalid signature",
			setupToken: func() string {
				validToken, _ := GenerateToken(1, "user")
				return validToken + "corrupted"
			},
			wantUserID: 0,
			wantRole:   "",
			wantErr:    true,
		},
		{
			name: "Token with future IssuedAt",
			setupToken: func() string {
				claims := Claims{
					UserID: 1,
					Role:   "user",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString(jwtKey)
				return tokenString
			},
			wantUserID: 0,
			wantRole:   "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken()
			claims, err := validator.ValidateToken(token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, tt.wantUserID, claims.UserID)
				assert.Equal(t, tt.wantRole, claims.Role)
			}
		})
	}
}

func TestTokenExpirationTimes(t *testing.T) {
	tests := []struct {
		name       string
		setupToken func() string
		wantErr    bool
	}{
		{
			name: "Token with no expiry",
			setupToken: func() string {
				claims := Claims{
					UserID: 1,
					Role:   "user",
					RegisteredClaims: jwt.RegisteredClaims{
						IssuedAt: jwt.NewNumericDate(time.Now()),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString(jwtKey)
				return tokenString
			},
			wantErr: false,
		},
		{
			name: "Token about to expire",
			setupToken: func() string {
				claims := Claims{
					UserID: 1,
					Role:   "user",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Second)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString(jwtKey)
				return tokenString
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken()
			claims, err := ValidateToken(token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
			}
		})
	}
}
