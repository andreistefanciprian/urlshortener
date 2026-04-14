// Package main is the entry point for the frontend2 service.
package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	logger := initLogger()

	backendHost := getEnvOrDefault("BACKEND_HOST", "api-gateway")
	backendPort := getEnvAsIntOrDefault("BACKEND_PORT", 8080)
	backendAddress := fmt.Sprintf("http://%s:%d", backendHost, backendPort)
	logger.Infof("Backend service address: %s", backendAddress)

	frontendApp := NewFrontendApp(logger, backendAddress)
	frontendPort := getEnvAsIntOrDefault("FRONTEND_PORT", 8090)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", frontendApp.home)
	mux.HandleFunc("POST /", frontendApp.home)
	mux.HandleFunc("GET /health", frontendApp.health)
	mux.HandleFunc("GET /ready", frontendApp.ready)
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("GET /.well-known/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", frontendPort),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Infof("Starting Frontend2 on port %d", frontendPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.WithError(err).Fatal("Failed to start HTTP server")
	}
}
