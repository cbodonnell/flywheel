package providers

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var _ AuthProvider = &FirebaseAuthProvider{}

type FirebaseAuthProvider struct {
	// app is the Firebase app
	app *firebase.App
	// auth is the Firebase Auth client
	auth *auth.Client
}

// NewFirebaseAuthProvider creates a new FirebaseAuthProvider
func NewFirebaseAuthProvider(ctx context.Context, projectID string, apiKey string) (*FirebaseAuthProvider, error) {
	opt := option.WithAPIKey(apiKey)
	cfg := &firebase.Config{
		ProjectID: projectID,
	}
	app, err := firebase.NewApp(ctx, cfg, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}

	auth, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %v", err)
	}

	return &FirebaseAuthProvider{
		app:  app,
		auth: auth,
	}, nil
}

// VerifyToken verifies a Firebase ID token
func (p *FirebaseAuthProvider) VerifyToken(ctx context.Context, idToken string) (*TokenClaims, error) {
	token, err := p.auth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("error verifying token: %v", err)
	}

	return &TokenClaims{
		UID: token.UID,
	}, nil
}
