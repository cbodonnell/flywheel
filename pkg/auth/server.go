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
	tls    *TLSConfig
}

type TLSConfig struct {
	CertFile string
	KeyFile  string
}

type NewAuthServerOptions struct {
	Port    int
	Handler handlers.AuthHandler
	TLS     *TLSConfig
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
		tls:    opts.TLS,
	}
}

// Start starts the AuthServer
func (s *AuthServer) Start() {
	var listenAndServe func() error
	if s.tls != nil {
		log.Info("Auth server listening on %s with TLS", s.server.Addr)
		listenAndServe = func() error {
			return s.server.ListenAndServeTLS(s.tls.CertFile, s.tls.KeyFile)
		}
	} else {
		log.Info("Auth server listening on %s", s.server.Addr)
		listenAndServe = s.server.ListenAndServe
	}
	if err := listenAndServe(); err != nil {
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
