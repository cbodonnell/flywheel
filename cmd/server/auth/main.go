package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cbodonnell/flywheel/pkg/auth"
	authhandlers "github.com/cbodonnell/flywheel/pkg/auth/handlers"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/version"
)

func main() {
	authPort := flag.Int("auth-port", 8080, "Auth server port")
	logLevel := flag.String("log-level", "info", "Log level")
	flag.Parse()

	parsedLogLevel, err := log.ParseLogLevel(*logLevel)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse log level: %v", err))
	}

	logger := log.New(os.Stdout, "", log.DefaultLoggerFlag, parsedLogLevel)
	log.SetDefaultLogger(logger)
	log.Info("Log level set to %s", parsedLogLevel)

	log.Info("Starting auth server version %s", version.Get())

	firebaseApiKey := os.Getenv("FLYWHEEL_FIREBASE_API_KEY")
	if firebaseApiKey == "" {
		panic("FLYWHEEL_FIREBASE_API_KEY environment variable must be set")
	}
	authServerOpts := auth.NewAuthServerOptions{
		Port:    *authPort,
		Handler: authhandlers.NewFirebaseAuthHandler(firebaseApiKey),
	}
	tlsCertFile := os.Getenv("FLYWHEEL_TLS_CERT_FILE")
	tlsKeyFile := os.Getenv("FLYWHEEL_TLS_KEY_FILE")
	if tlsCertFile != "" && tlsKeyFile != "" {
		authServerOpts.TLS = &auth.TLSConfig{
			CertFile: tlsCertFile,
			KeyFile:  tlsKeyFile,
		}
	}
	authServer := auth.NewAuthServer(authServerOpts)
	authServer.Start()
}
