package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/cbodonnell/flywheel/pkg/auth/handlers"
	"github.com/cbodonnell/flywheel/pkg/log"
)

type AuthServer struct {
	server *http.Server
}

type NewAuthServerOptions struct {
	Port    int
	Handler handlers.AuthHandler
}

// NewAuthServer creates a new http.Server for handling authentication requests
func NewAuthServer(opts NewAuthServerOptions) *AuthServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", opts.Handler.HandleRegister())
	mux.HandleFunc("/login", opts.Handler.HandleLogin())
	mux.HandleFunc("/refresh", opts.Handler.HandleRefresh())
	mux.HandleFunc("/delete", opts.Handler.HandleDelete())
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", opts.Port),
		Handler: mux,
	}
	return &AuthServer{
		server: server,
	}
}

// Start starts the AuthServer
func (s *AuthServer) Start() {
	log.Info("Auth server listening on %s", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Info("Auth server closed")
			return
		}
		log.Error("Auth server error: %v", err)
	}
}

// Stop stops the AuthServer
func (s *AuthServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
