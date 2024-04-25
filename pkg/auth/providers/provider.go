package providers

import "context"

type AuthProvider interface {
	VerifyToken(ctx context.Context, idToken string) (*TokenClaims, error)
}

type TokenClaims struct {
	UID string `json:"uid"`
}
