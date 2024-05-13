package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	authproviders "github.com/cbodonnell/flywheel/pkg/auth/providers"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/repositories"
)

type ContextKey int

const (
	// UserContextKey is the key used to store the user in the request context
	UserContextKey ContextKey = iota
)

func NewAuthMiddleware(authProvider authproviders.AuthProvider, repository repositories.Repository) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearerToken, err := parseBearerToken(r)
			if err != nil {
				log.Error("failed to parse bearer token: %v", err)
				http.Error(w, "failed to parse bearer token", http.StatusUnauthorized)
				return
			}

			token, err := authProvider.VerifyToken(r.Context(), bearerToken)
			if err != nil {
				log.Error("failed to verify ID token: %v", err)
				http.Error(w, "failed to verify ID token", http.StatusUnauthorized)
				return
			}

			user, err := repository.CreateUser(r.Context(), token.UID)
			if err != nil {
				log.Error("failed to create user: %v", err)
				http.Error(w, "failed to create user", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// parseBearerToken parses the bearer token from the Authorization header
func parseBearerToken(r *http.Request) (string, error) {
	// Get the Authorization header value
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is missing")
	}

	// Check if the Authorization header has the Bearer scheme
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("invalid Authorization header format")
	}

	// Return the token part
	return parts[1], nil
}
