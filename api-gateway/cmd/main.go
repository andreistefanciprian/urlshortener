package main

import (
	"fmt"
	"net/http"
	"time"

	ugen "github.com/andreistefanciprian/urlshortener/api-gateway/url-gen/proto"
	uread "github.com/andreistefanciprian/urlshortener/api-gateway/url-read/proto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Initialize Prometheus metrics
	initMetrics()

	// Initialize the logger
	logger := initLogger()

	// Connect to URL Generation service
	urlGenHost := getEnvOrDefault("URL_GEN_HOST", "url-gen")
	urlGenPort := getEnvAsIntOrDefault("URL_GEN_PORT", 50052)
	urlGenAddress := fmt.Sprintf("%s:%d", urlGenHost, urlGenPort)
	urlGenConn, err := grpc.NewClient(urlGenAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to URL Generation service")
	}
	defer func() {
		if err := urlGenConn.Close(); err != nil {
			logger.WithError(err).Error("Failed to close gRPC connection")
		}
	}()
	urlGenClient := ugen.NewURLGeneratorClient(urlGenConn)
	logger.Infof("Connected to URL Generation service at %s", urlGenAddress)

	// Connect to URL Reading service
	urlReadHost := getEnvOrDefault("URL_READ_HOST", "url-read")
	urlReadPort := getEnvAsIntOrDefault("URL_READ_PORT", 50053)
	urlReadAddress := fmt.Sprintf("%s:%d", urlReadHost, urlReadPort)
	urlReadConn, err := grpc.NewClient(urlReadAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to URL Reading service")
	}
	defer func() {
		if err := urlReadConn.Close(); err != nil {
			logger.WithError(err).Error("Failed to close gRPC connection")
		}
	}()
	urlReadClient := uread.NewURLReaderClient(urlReadConn)
	logger.Infof("Connected to URL Reading service at %s", urlReadAddress)

	// Initialize HTTP handler
	urlGateway := NewURLGateway(logger, urlGenClient, urlReadClient)

	// Set up metrics server
	metricsPort := getEnvAsIntOrDefault("METRICS_PORT", 9090)
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsSrv := &http.Server{
		Addr:              fmt.Sprintf(":%d", metricsPort),
		Handler:           metricsMux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	// Start metrics server in a separate goroutine
	go func() {
		logger.Infof("Starting metrics server on port %d", metricsPort)
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start metrics server")
		}
	}()

	// Set up main server
	apiPort := getEnvAsIntOrDefault("API_GATEWAY_PORT", 8080)
	api := http.NewServeMux()
	api.HandleFunc("POST /create", urlGateway.CreateShortURL)
	api.HandleFunc("GET /{shortCode}", urlGateway.GetLongURL)
	api.HandleFunc("DELETE /{shortCode}", urlGateway.DeleteShortURL)
	api.HandleFunc("GET /", urlGateway.GetAllURLs)
	apiSrv := &http.Server{
		Addr:              fmt.Sprintf(":%d", apiPort),
		Handler:           api,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	// Start main HTTP server in a separate go routine
	go func() {
		logger.Infof("Starting API Gateway on port %d", apiPort)
		if err := apiSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Handle graceful shutdown
	handleGracefulShutdown(logger, map[string]*http.Server{
		"metrics": metricsSrv,
		"api":     apiSrv,
	})
}
