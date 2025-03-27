package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/25x8/ya-prakt-6-sprint/internal/gophermart/repository"
	"github.com/golang-jwt/jwt/v4"
)

type contextKey string

const (
	// UserIDKey is the key for user ID in the request context
	UserIDKey contextKey = "userID"
	// Authentication-related constants
	jwtExpirationTime = 24 * time.Hour
	authCookieName    = "auth_token"
	bearerSchema      = "Bearer "
)

// JWTConfig contains configuration for JWT authentication
type JWTConfig struct {
	SecretKey string
	Repo      repository.Repository
}

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken generates a JWT token for a user
func GenerateToken(userID int64, secretKey string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtExpirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// AuthMiddleware creates middleware that checks if the user is authenticated
func AuthMiddleware(jwtConfig *JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header or cookie
			tokenString := extractToken(r)
			if tokenString == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Parse token
			token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(jwtConfig.SecretKey), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Extract user ID from claims
			claims, ok := token.Claims.(*JWTClaims)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Verify that user exists in database
			ctx := r.Context()
			user, err := jwtConfig.Repo.GetUserByID(ctx, claims.UserID)
			if err != nil || user == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add user ID to request context
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken extracts JWT token from Authorization header or cookie
func extractToken(r *http.Request) string {
	// Try from Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, bearerSchema) {
		return strings.TrimPrefix(authHeader, bearerSchema)
	}

	// Try from cookie
	cookie, err := r.Cookie(authCookieName)
	if err == nil {
		return cookie.Value
	}

	return ""
}

// SetAuthCookie sets authentication cookie
func SetAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(jwtExpirationTime.Seconds()),
	})
}

// GetUserID extracts user ID from request context
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}
