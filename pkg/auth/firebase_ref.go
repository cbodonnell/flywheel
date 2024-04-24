package auth

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

func NewFirebaseApp(ctx context.Context, credentialsPath string) (*firebase.App, error) {
	// opt := option.WithAPIKey(apiKey)
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}
	return app, nil
}

func NewFirebaseAuthClient(ctx context.Context, app *firebase.App) (*auth.Client, error) {
	auth, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %v", err)
	}
	return auth, nil
}

func VerifyToken(ctx context.Context, client *auth.Client, idToken string) (*auth.Token, error) {
	token, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("error verifying token: %v", err)
	}
	return token, nil
}

func RevokedToken(ctx context.Context, client *auth.Client, uid string) error {
	if err := client.RevokeRefreshTokens(ctx, uid); err != nil {
		return fmt.Errorf("error revoking refresh tokens: %v", err)
	}
	return nil
}

func GetUser(ctx context.Context, client *auth.Client, uid string) (*auth.UserRecord, error) {
	user, err := client.GetUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("error fetching user data: %v", err)
	}
	return user, nil
}

func CreateUser(ctx context.Context, client *auth.Client, email, password string) (*auth.UserRecord, error) {
	params := (&auth.UserToCreate{}).
		Email(email).
		Password(password)
	user, err := client.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v", err)
	}
	return user, nil
}
