package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/cbodonnell/flywheel/pkg/auth/handlers"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/middleware"
	"github.com/gorilla/mux"
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
	Port        int
	AllowOrigin string
	Handler     handlers.AuthHandler
	TLS         *TLSConfig
}

// NewAuthServer creates a new http.Server for handling authentication requests
func NewAuthServer(opts NewAuthServerOptions) *AuthServer {
	r := mux.NewRouter()

	r.HandleFunc("/register", opts.Handler.HandleRegister()).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/login", opts.Handler.HandleLogin()).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/refresh", opts.Handler.HandleRefresh()).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/delete", opts.Handler.HandleDelete()).Methods(http.MethodPost, http.MethodOptions)

	corsMiddleware := middleware.NewCORSMiddleware(opts.AllowOrigin)
	optionsMiddleware := middleware.NewOptionsMiddleware()
	r.Use(corsMiddleware, optionsMiddleware)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", opts.Port),
		Handler: r,
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
