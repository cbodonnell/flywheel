package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/cbodonnell/flywheel/pkg/api"
	authproviders "github.com/cbodonnell/flywheel/pkg/auth/providers"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/repositories"
	"github.com/cbodonnell/flywheel/pkg/version"
)

func main() {
	port := flag.Int("port", 9090, "port to listen on")
	allowOrigin := flag.String("allow-origin", "localhost", "comma-separated list of allowed origins")
	logLevel := flag.String("log-level", "info", "Log level")
	flag.Parse()

	parsedLogLevel, err := log.ParseLogLevel(*logLevel)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse log level: %v", err))
	}

	logger := log.New(os.Stdout, "", log.DefaultLoggerFlag, parsedLogLevel)
	log.SetDefaultLogger(logger)
	log.Info("Log level set to %s", parsedLogLevel)

	log.Info("Starting api server version %s", version.Get())
	ctx := context.Background()

	firebaseProjectID := os.Getenv("FLYWHEEL_FIREBASE_PROJECT_ID")
	if firebaseProjectID == "" {
		panic("FLYWHEEL_FIREBASE_PROJECT_ID environment variable must be set")
	}
	authProvider, err := authproviders.NewFirebaseAuthProvider(ctx, firebaseProjectID)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Firebase auth provider: %v", err))
	}

	connStr := os.Getenv("FLYWHEEL_DATABASE_URL")
	if connStr == "" {
		connStr = "sqlite://flywheel.db"
	}

	u, err := url.Parse(connStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse connection string: %v", err))
	}

	var repository repositories.Repository
	switch u.Scheme {
	case "sqlite":
		repository, err = repositories.NewSQLiteRepository(ctx, u.Host, "./migrations/sqlite")
		if err != nil {
			panic(fmt.Sprintf("Failed to create SQLite repository: %v", err))
		}
	case "postgresql":
		repository, err = repositories.NewPostgresRepository(ctx, u.String())
		if err != nil {
			panic(fmt.Sprintf("Failed to create Postgres repository: %v", err))
		}
	default:
		panic(fmt.Sprintf("Unknown database type %s", u.Scheme))
	}
	defer repository.Close(ctx)

	apiServerOpts := api.NewAPIServerOptions{
		Port:         *port,
		AllowOrigin:  *allowOrigin,
		AuthProvider: authProvider,
		Repository:   repository,
	}
	tlsCertFile := os.Getenv("FLYWHEEL_API_TLS_CERT_FILE")
	tlsKeyFile := os.Getenv("FLYWHEEL_API_TLS_KEY_FILE")
	if tlsCertFile != "" && tlsKeyFile != "" {
		apiServerOpts.TLS = &api.TLSConfig{
			CertFile: tlsCertFile,
			KeyFile:  tlsKeyFile,
		}
	}
	server := api.NewAPIServer(apiServerOpts)
	go server.Start()

	interrupt := make(chan os.Signal, 1)
	<-interrupt
	if err := server.Stop(ctx); err != nil {
		log.Error("Failed to stop server: %v", err)
	}
}
