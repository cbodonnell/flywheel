package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/cbodonnell/flywheel/pkg/api/handlers"
	authproviders "github.com/cbodonnell/flywheel/pkg/auth/providers"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/middleware"
	"github.com/cbodonnell/flywheel/pkg/repositories"
	"github.com/gorilla/mux"
)

type APIServer struct {
	server *http.Server
	tls    *TLSConfig
}

type TLSConfig struct {
	CertFile string
	KeyFile  string
}

type NewAPIServerOptions struct {
	Port         int
	AllowOrigin  string
	TLS          *TLSConfig
	AuthProvider authproviders.AuthProvider
	Repository   repositories.Repository
}

// NewAPIServer creates a new http.Server for handling API requests
func NewAPIServer(opts NewAPIServerOptions) *APIServer {
	r := mux.NewRouter()

	r.HandleFunc("/characters", handlers.HandleListCharacters(opts.Repository)).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/characters", handlers.HandleCreateCharacter(opts.Repository)).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/characters/{characterID}", handlers.HandleDeleteCharacter(opts.Repository)).Methods(http.MethodDelete, http.MethodOptions)

	corsMiddleware := middleware.NewCORSMiddleware(opts.AllowOrigin)
	optionsMiddleware := middleware.NewOptionsMiddleware()
	authMiddleware := middleware.NewAuthMiddleware(opts.AuthProvider, opts.Repository)
	r.Use(corsMiddleware, optionsMiddleware, authMiddleware)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", opts.Port),
		Handler: r,
	}
	return &APIServer{
		server: server,
		tls:    opts.TLS,
	}
}

// Start starts the APIServer
func (s *APIServer) Start() {
	var listenAndServe func() error
	if s.tls != nil {
		log.Info("API server listening on %s with TLS", s.server.Addr)
		listenAndServe = func() error {
			return s.server.ListenAndServeTLS(s.tls.CertFile, s.tls.KeyFile)
		}
	} else {
		log.Info("API server listening on %s", s.server.Addr)
		listenAndServe = s.server.ListenAndServe
	}
	if err := listenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Info("API server closed")
			return
		}
		log.Error("API server error: %v", err)
	}
}

// Stop stops the APIServer
func (s *APIServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
