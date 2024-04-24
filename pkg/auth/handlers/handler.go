package handlers

import "net/http"

// AuthHandler is an interface for handling authentication requests
type AuthHandler interface {
	HandleRegister() func(w http.ResponseWriter, r *http.Request)
	HandleLogin() func(w http.ResponseWriter, r *http.Request)
	HandleRefresh() func(w http.ResponseWriter, r *http.Request)
	HandleDelete() func(w http.ResponseWriter, r *http.Request)
}
