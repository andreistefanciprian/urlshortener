package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {

	// Initialize the logger
	logger := initLogger()

	// APi Gateway environment variables
	backendHost := getEnvOrDefault("BACKEND_HOST", "api-gateway")
	backendPort := getEnvAsIntOrDefault("BACKEND_PORT", 8080)
	backendAddress := fmt.Sprintf("http://%s:%d", backendHost, backendPort)
	logger.Infof("Backend service address: %s", backendAddress)

	// Set up main server
	frontendApp := NewFrontendApp(logger, backendAddress)
	frontendPort := getEnvAsIntOrDefault("FRONTEND_PORT", 8090)
	frontend := http.NewServeMux()
	frontend.HandleFunc("GET /", frontendApp.home)
	frontend.HandleFunc("POST /", frontendApp.home)
	// Add favicon handler to prevent unnecessary requests
	frontend.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	// Add handler for Chrome DevTools and other well-known paths
	frontend.HandleFunc("GET /.well-known/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	apiSrv := &http.Server{
		Addr:              fmt.Sprintf(":%d", frontendPort),
		Handler:           frontend,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Infof("Starting Frontend on port %d", frontendPort)
	if err := apiSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.WithError(err).Fatal("Failed to start HTTP server")
	}

}
