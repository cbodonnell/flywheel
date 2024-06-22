package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/cbodonnell/flywheel/pkg/log"
)

var _ AuthHandler = &FirebaseAuthHandler{}

// FirebaseAuthHandler implements AuthHandler using Firebase Auth REST API
type FirebaseAuthHandler struct {
	apiKey string
}

// NewFirebaseAuthHandler creates a new instance of FirebaseAuthHandler
func NewFirebaseAuthHandler(apiKey string) *FirebaseAuthHandler {
	return &FirebaseAuthHandler{
		apiKey: apiKey,
	}
}

// ErrorResponseBody is the response body for an error
// https://firebase.google.com/docs/reference/rest/auth#section-error-format
type ErrorResponseBody struct {
	Error struct {
		Code    int                  `json:"code"`
		Message ErrorResponseMessage `json:"message"`
		Errors  []struct {
			Message string `json:"message"`
			Domain  string `json:"domain"`
			Reason  string `json:"reason"`
		} `json:"errors"`
	} `json:"error"`
}

type ErrorResponseMessage string

const (
	ErrorEmailExists             ErrorResponseMessage = "EMAIL_EXISTS"
	ErrorOperationNotAllowed     ErrorResponseMessage = "OPERATION_NOT_ALLOWED"
	ErrorTooManyAttempts         ErrorResponseMessage = "TOO_MANY_ATTEMPTS_TRY_LATER"
	ErrorInvalidEmail            ErrorResponseMessage = "INVALID_EMAIL"
	ErrorInvalidLoginCredentials ErrorResponseMessage = "INVALID_LOGIN_CREDENTIALS"
	ErrorTokenExpired            ErrorResponseMessage = "TOKEN_EXPIRED"
	ErrorInvalidIDToken          ErrorResponseMessage = "INVALID_ID_TOKEN"
	ErrorUserNotFound            ErrorResponseMessage = "USER_NOT_FOUND"
	ErrorWeakPassword            ErrorResponseMessage = "WEAK_PASSWORD : Password should be at least 6 characters"
)

// RegisterRequestBody is the request body for the register endpoint
type RegisterRequestBody struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ReturnSecureToken bool   `json:"returnSecureToken"`
}

// RegisterResponseBody is the response body for the register endpoint
type RegisterResponseBody struct {
	IDToken      string `json:"idToken"`
	Email        string `json:"email"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
	LocalID      string `json:"localId"`
}

// HandleRegister handles requests to the register endpoint
// https://firebase.google.com/docs/reference/rest/auth#section-create-email-password
func (s *FirebaseAuthHandler) HandleRegister() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		password := r.FormValue("password")

		if email == "" {
			http.Error(w, "Missing email", http.StatusBadRequest)
			return
		}
		if password == "" {
			http.Error(w, "Missing password", http.StatusBadRequest)
			return
		}

		requestPayload := &RegisterRequestBody{
			Email:             email,
			Password:          password,
			ReturnSecureToken: true,
		}

		body := bytes.NewBuffer(nil)
		if err := json.NewEncoder(body).Encode(requestPayload); err != nil {
			log.Error("error encoding request body: %v", err)
			http.Error(w, "error encoding request body", http.StatusInternalServerError)
			return
		}

		req, err := http.NewRequest("POST", "https://identitytoolkit.googleapis.com/v1/accounts:signUp?key="+s.apiKey, body)
		if err != nil {
			log.Error("error creating request: %v", err)
			http.Error(w, "error creating request", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := http.DefaultClient
		resp, err := client.Do(req)
		if err != nil {
			log.Error("error sending request: %v", err)
			http.Error(w, "error sending request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Error("error response status: %s", resp.Status)
			errorResponse := &ErrorResponseBody{}
			if err := json.NewDecoder(resp.Body).Decode(errorResponse); err != nil {
				log.Error("failed to decode error response: %v", err)
				http.Error(w, "failed to decode error response", http.StatusInternalServerError)
				return
			}

			switch errorResponse.Error.Message {
			case ErrorInvalidEmail:
				http.Error(w, "Invalid email", http.StatusBadRequest)
				return
			case ErrorWeakPassword:
				http.Error(w, "Password should be at least 6 characters", http.StatusBadRequest)
				return
			case ErrorEmailExists:
				http.Error(w, "Email already exists", http.StatusBadRequest)
				return
			case ErrorOperationNotAllowed:
				http.Error(w, "Operation not allowed", http.StatusBadRequest)
				return
			case ErrorTooManyAttempts:
				http.Error(w, "Too many attempts, try again later", http.StatusBadRequest)
				return
			}

			log.Error("unhandled error response message: %s", errorResponse.Error.Message)
			http.Error(w, "Failed to register", http.StatusInternalServerError)
			return
		}

		responsePayload := &RegisterResponseBody{}
		if err := json.NewDecoder(resp.Body).Decode(responsePayload); err != nil {
			log.Error("error decoding response: %v", err)
			http.Error(w, "error decoding response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
			log.Error("error encoding response: %v", err)
			http.Error(w, "error encoding response", http.StatusInternalServerError)
			return
		}
	}
}

// LoginRequestBody is the request body for the login endpoint
type LoginRequestBody struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ReturnSecureToken bool   `json:"returnSecureToken"`
}

// LoginResponseBody is the response body for the login endpoint
type LoginResponseBody struct {
	IDToken      string `json:"idToken"`
	Email        string `json:"email"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
	LocalID      string `json:"localId"`
	Registered   bool   `json:"registered"`
}

// HandleLogin handles requests to the login endpoint
// https://firebase.google.com/docs/reference/rest/auth#section-sign-in-email-password
func (s *FirebaseAuthHandler) HandleLogin() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		password := r.FormValue("password")

		if email == "" {
			http.Error(w, "Missing email", http.StatusBadRequest)
			return
		}
		if password == "" {
			http.Error(w, "Missing password", http.StatusBadRequest)
			return
		}

		requestPayload := &LoginRequestBody{
			Email:             email,
			Password:          password,
			ReturnSecureToken: true,
		}

		body := bytes.NewBuffer(nil)
		if err := json.NewEncoder(body).Encode(requestPayload); err != nil {
			log.Error("error encoding request body: %v", err)
			http.Error(w, "error encoding request body", http.StatusInternalServerError)
			return
		}

		req, err := http.NewRequest("POST", "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key="+s.apiKey, body)
		if err != nil {
			log.Error("error creating request: %v", err)
			http.Error(w, "error creating request", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := http.DefaultClient
		resp, err := client.Do(req)
		if err != nil {
			log.Error("error sending request: %v", err)
			http.Error(w, "error sending request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Error("error response status: %s", resp.Status)
			errorResponse := &ErrorResponseBody{}
			if err := json.NewDecoder(resp.Body).Decode(errorResponse); err != nil {
				log.Error("failed to decode error response: %v", err)
				http.Error(w, "failed to decode error response", http.StatusInternalServerError)
				return
			}

			switch errorResponse.Error.Message {
			case ErrorInvalidEmail:
				http.Error(w, "Invalid email", http.StatusBadRequest)
				return
			case ErrorInvalidLoginCredentials:
				http.Error(w, "Invalid credentials", http.StatusBadRequest)
				return
			}

			log.Error("unhandled error response message: %s", errorResponse.Error.Message)
			http.Error(w, "Failed to login", http.StatusInternalServerError)
			return
		}

		responsePayload := &LoginResponseBody{}
		if err := json.NewDecoder(resp.Body).Decode(responsePayload); err != nil {
			log.Error("error decoding response: %v", err)
			http.Error(w, "error decoding response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*") // TODO: CORS :/
		if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
			log.Error("error encoding response: %v", err)
			http.Error(w, "error encoding response", http.StatusInternalServerError)
			return
		}
	}
}

// RefreshRequestBody is the request body for the refresh endpoint
type RefreshRequestBody struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshResponseBody is the response body for the refresh endpoint
type RefreshResponseBody struct {
	ExpiresIn    string `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	UserID       string `json:"user_id"`
	ProjectID    string `json:"project_id"`
}

// HandleRefresh handles requests to the refresh endpoint
// https://firebase.google.com/docs/reference/rest/auth#section-refresh-token
func (s *FirebaseAuthHandler) HandleRefresh() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		refreshToken := r.FormValue("refreshToken")

		if refreshToken == "" {
			http.Error(w, "Missing refresh token", http.StatusBadRequest)
			return
		}

		requestPayload := &RefreshRequestBody{
			GrantType:    "refresh_token",
			RefreshToken: refreshToken,
		}

		body := bytes.NewBuffer(nil)
		if err := json.NewEncoder(body).Encode(requestPayload); err != nil {
			log.Error("error encoding request body: %v", err)
			http.Error(w, "error encoding request body", http.StatusInternalServerError)
			return
		}

		req, err := http.NewRequest("POST", "https://securetoken.googleapis.com/v1/token?key="+s.apiKey, body)
		if err != nil {
			log.Error("error creating request: %v", err)
			http.Error(w, "error creating request", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := http.DefaultClient
		resp, err := client.Do(req)
		if err != nil {
			log.Error("error sending request: %v", err)
			http.Error(w, "error sending request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Error("error response status: %s", resp.Status)
			errorResponse := &ErrorResponseBody{}
			if err := json.NewDecoder(resp.Body).Decode(errorResponse); err != nil {
				log.Error("failed to decode error response: %v", err)
				http.Error(w, "failed to decode error response", http.StatusInternalServerError)
				return
			}

			switch errorResponse.Error.Message {
			case ErrorTokenExpired:
				http.Error(w, "Token expired", http.StatusBadRequest)
				return
			}

			log.Error("unhandled error response message: %s", errorResponse.Error.Message)
			http.Error(w, "Failed to refresh", http.StatusInternalServerError)
			return
		}

		responsePayload := &RefreshResponseBody{}
		if err := json.NewDecoder(resp.Body).Decode(responsePayload); err != nil {
			log.Error("error decoding response: %v", err)
			http.Error(w, "error decoding response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
			log.Error("error encoding response: %v", err)
			http.Error(w, "error encoding response", http.StatusInternalServerError)
			return
		}
	}
}

// DeleteRequestBody is the request body for the delete endpoint
type DeleteRequestBody struct {
	IDToken string `json:"idToken"`
}

// HandleDelete handles requests to the delete endpoint
// https://firebase.google.com/docs/reference/rest/auth#section-delete-account
func (s *FirebaseAuthHandler) HandleDelete() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idToken := r.FormValue("idToken")

		if idToken == "" {
			http.Error(w, "Missing ID token", http.StatusBadRequest)
			return
		}

		requestPayload := &DeleteRequestBody{
			IDToken: idToken,
		}

		body := bytes.NewBuffer(nil)
		if err := json.NewEncoder(body).Encode(requestPayload); err != nil {
			log.Error("error encoding request body: %v", err)
			http.Error(w, "error encoding request body", http.StatusInternalServerError)
			return
		}

		req, err := http.NewRequest("POST", "https://identitytoolkit.googleapis.com/v1/accounts:delete?key="+s.apiKey, body)
		if err != nil {
			log.Error("error creating request: %v", err)
			http.Error(w, "error creating request", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := http.DefaultClient
		resp, err := client.Do(req)
		if err != nil {
			log.Error("error sending request: %v", err)
			http.Error(w, "error sending request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Error("error response status: %s", resp.Status)
			errorResponse := &ErrorResponseBody{}
			if err := json.NewDecoder(resp.Body).Decode(errorResponse); err != nil {
				log.Error("failed to decode error response: %v", err)
				http.Error(w, "failed to decode error response", http.StatusInternalServerError)
				return
			}

			switch errorResponse.Error.Message {
			case ErrorInvalidIDToken:
				http.Error(w, "Invalid ID token", http.StatusBadRequest)
			case ErrorUserNotFound:
				http.Error(w, "User not found", http.StatusBadRequest)
				return
			}

			log.Error("unhandled error response message: %s", errorResponse.Error.Message)
			http.Error(w, "Failed to delete", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
