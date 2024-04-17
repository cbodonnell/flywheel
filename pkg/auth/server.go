package auth

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

func NewFirebaseApp(credentialsPath string) (*firebase.App, error) {
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}
	return app, nil
}

func NewFirebaseAuthClient(app *firebase.App) (*auth.Client, error) {
	auth, err := app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %v", err)
	}
	return auth, nil
}

func VerifyToken(client *auth.Client, idToken string) (*auth.Token, error) {
	token, err := client.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		return nil, fmt.Errorf("error verifying token: %v", err)
	}
	return token, nil
}

func RevokedToken(client *auth.Client, uid string) error {
	if err := client.RevokeRefreshTokens(context.Background(), uid); err != nil {
		return fmt.Errorf("error revoking refresh tokens: %v", err)
	}
	return nil
}

func GetUser(client *auth.Client, uid string) (*auth.UserRecord, error) {
	user, err := client.GetUser(context.Background(), uid)
	if err != nil {
		return nil, fmt.Errorf("error fetching user data: %v", err)
	}
	return user, nil
}

func CreateUser(client *auth.Client, email, password string) (*auth.UserRecord, error) {
	params := (&auth.UserToCreate{}).
		Email(email).
		Password(password)
	user, err := client.CreateUser(context.Background(), params)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v", err)
	}
	return user, nil
}
