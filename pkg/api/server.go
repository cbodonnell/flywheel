package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/cbodonnell/flywheel/pkg/api/handlers"
	"github.com/cbodonnell/flywheel/pkg/api/middleware"
	authproviders "github.com/cbodonnell/flywheel/pkg/auth/providers"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/repositories"
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
	TLS          *TLSConfig
	AuthProvider authproviders.AuthProvider
	Repository   repositories.Repository
}

// NewAPIServer creates a new http.Server for handling API requests
func NewAPIServer(opts NewAPIServerOptions) *APIServer {
	authMiddleware := middleware.NewAuthMiddleware(opts.AuthProvider, opts.Repository)

	mux := http.NewServeMux()
	mux.Handle("/characters", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		switch r.Method {
		case http.MethodOptions:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			handlers.HandleListCharacters(opts.Repository)(w, r)
		case http.MethodPost:
			handlers.HandleCreateCharacter(opts.Repository)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/characters/{characterID}", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		switch r.Method {
		case http.MethodDelete:
			handlers.HandleDeleteCharacter(opts.Repository)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", opts.Port),
		Handler: mux,
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
