package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	ugen "github.com/andreistefanciprian/urlshortener/api-gateway/url-gen/proto"
	uread "github.com/andreistefanciprian/urlshortener/api-gateway/url-read/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Prometheus metrics
var (
	createShortURLTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "create_short_url_total",
			Help: "Total number of short URL creation requests",
		},
		[]string{"status"}, // Labels: success, error, invalid_url, invalid_expiration
	)

	getLongURLTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "get_long_url_total",
			Help: "Total number of long URL retrieval requests",
		},
		[]string{"status"}, // Labels: success, error, not_found, expired
	)
)

func init() {
	// Register Prometheus metrics
	prometheus.MustRegister(createShortURLTotal)
	prometheus.MustRegister(getLongURLTotal)
}

// getEnvAsIntOrDefault returns environment variable as int or default if not set or invalid
func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {

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
	handleGracefulShutdown(logger, metricsSrv, apiSrv)
}

// handleGracefulShutdown waits for interrupt signals and gracefully shuts down the provided servers
func handleGracefulShutdown(logger *logrus.Logger, servers ...*http.Server) {
	// Wait for interrupt signal to gracefully shutdown the servers
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down servers...")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown all servers
	for i, server := range servers {
		serverName := fmt.Sprintf("Server %d", i+1)
		switch i {
		case 0:
			serverName = "Metrics server"
		case 1:
			serverName = "API server"
		}

		if err := server.Shutdown(ctx); err != nil {
			logger.WithError(err).Errorf("%s forced to shutdown", serverName)
		} else {
			logger.Infof("%s shutdown gracefully", serverName)
		}
	}

	logger.Info("Servers shutdown complete")
}

type URLHandler interface {
	// Define methods for handling HTTP requests
	CreateShortURL(w http.ResponseWriter, r *http.Request)
	GetLongURL(w http.ResponseWriter, r *http.Request)
}

type URLGateway struct {
	logger        *logrus.Logger
	urlGenClient  ugen.URLGeneratorClient // gRPC URLGeneratorClient
	urlReadClient uread.URLReaderClient   // gRPC URLReaderClient
}

func NewURLGateway(logger *logrus.Logger, urlGenClient ugen.URLGeneratorClient, urlReadClient uread.URLReaderClient) *URLGateway {
	return &URLGateway{
		logger:        logger,
		urlGenClient:  urlGenClient,
		urlReadClient: urlReadClient,
	}
}

func (h *URLGateway) GetLongURL(w http.ResponseWriter, r *http.Request) {
	// Use request context for future enhancements such as logging, timeouts, tracing, etc.
	ctx := r.Context()

	// Extract short URL code from URL path
	shortUrlCode := r.PathValue("shortCode")

	// Validate required fields
	if err := validateShortURL(shortUrlCode); err != nil {
		getLongURLTotal.WithLabelValues("error").Inc()
		h.logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create gRPC request
	grpcReq := &uread.LongURLRequest{
		ShortUrl: shortUrlCode,
	}

	// Call gRPC service
	response, err := h.urlReadClient.GetLongURL(ctx, grpcReq)
	if err != nil {
		// Check if it's a "not found" error (which could be an expired/deleted URL)
		if strings.Contains(err.Error(), "not found") {
			getLongURLTotal.WithLabelValues("not_found").Inc()
			h.logger.Infof("Short URL %s not found (may be expired or deleted)", shortUrlCode)
			http.Error(w, "Short URL not found", http.StatusNotFound)
			return
		}
		getLongURLTotal.WithLabelValues("error").Inc()
		h.logger.Errorf("gRPC call to GetLongURL failed: %v", err)
		http.Error(w, "Failed to retrieve long URL", http.StatusInternalServerError)
		return
	}

	// Check if response.Expiration is set and if the URL has expired
	if response.Expiration != nil {
		if response.Expiration.AsTime().Before(time.Now()) {
			getLongURLTotal.WithLabelValues("expired").Inc()
			h.logger.Infof("Short URL %s has expired at %v", shortUrlCode, response.Expiration.AsTime())
			http.Error(w, "Short URL has expired", http.StatusGone)
			return
		}
	}

	// Success - increment counter and redirect
	getLongURLTotal.WithLabelValues("success").Inc()

	// Redirect with 302 to the original long URL
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Expires", "0")
	http.Redirect(w, r, response.LongUrl, http.StatusFound)
}

const (
	shortUrlCodeLength     = 7
	URLShortenerDomainName = "l.it"
)

func validateShortURL(shortURL string) error {
	// Basic validation: check if the short URL is not empty and has a valid format
	if shortURL == "" {
		return fmt.Errorf("short URL is empty")
	}

	if len(shortURL) != shortUrlCodeLength {
		return fmt.Errorf("invalid short URL length")
	}

	// Further validation logic can be added here (e.g., regex check)
	return nil
}

type CreateShortURLRequest struct {
	LongUrl   string `json:"longUrl"`
	ExpiresIn int    `json:"expiresIn"` // Expiration time in days
}

func (u *CreateShortURLRequest) validateExpiration() error {

	// Validate expiration time
	if u.ExpiresIn < 0 {
		return fmt.Errorf("expiration time cannot be negative")
	}

	// Set default expiration if not provided
	if u.ExpiresIn == 0 {
		u.ExpiresIn = 7 // Default to 7 days if not provided
	}

	return nil
}

func (u *CreateShortURLRequest) validateURL() error {
	// Parse and validate the URL using net/url ParseRequestURI according to RFC3986 for request URIs
	parsedURL, err := url.ParseRequestURI(u.LongUrl)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}

	// Validate URL scheme - only allow http and https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme, got: %s", parsedURL.Scheme)
	}

	// Validate that the URL has a host (catch https:// no Host URLs)
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	// Normalize the URL for consistent storage and deduplication
	// Convert scheme and host to lowercase for proper normalization
	parsedURL.Scheme = strings.ToLower(parsedURL.Scheme)
	parsedURL.Host = strings.ToLower(parsedURL.Host)

	// Remove default ports for cleaner URLs
	if parsedURL.Scheme == "https" && strings.HasSuffix(parsedURL.Host, ":443") {
		parsedURL.Host = strings.TrimSuffix(parsedURL.Host, ":443")
	} else if parsedURL.Scheme == "http" && strings.HasSuffix(parsedURL.Host, ":80") {
		parsedURL.Host = strings.TrimSuffix(parsedURL.Host, ":80")
	}

	// Set the normalized URL back to the request
	u.LongUrl = parsedURL.String()
	return nil
}

func (h *URLGateway) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	// Use request context for future enhancements such as logging, timeouts, tracing, etc.
	ctx := r.Context()

	// Parse request body
	var req CreateShortURLRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		createShortURLTotal.WithLabelValues("error").Inc()
		h.logger.Errorf("Failed to decode JSON body: %v", err)
		http.Error(w, "Failed to decode JSON body", http.StatusBadRequest)
		return
	}

	// Validate required fields and set defaults
	if err := req.validateURL(); err != nil {
		createShortURLTotal.WithLabelValues("invalid_url").Inc()
		h.logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.validateExpiration(); err != nil {
		createShortURLTotal.WithLabelValues("invalid_expiration").Inc()
		h.logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Log the incoming request
	h.logger.WithContext(ctx).WithFields(logrus.Fields{
		"longUrl":   req.LongUrl,
		"expiresIn": req.ExpiresIn,
	}).Info("Received CreateShortURL request")

	// Create gRPC request with the validated and normalized URL
	grpcReq := &ugen.ShortURLRequest{
		LongUrl:    req.LongUrl, // This is now validated and normalized
		Expiration: timestamppb.New(time.Now().Add(time.Duration(req.ExpiresIn) * 24 * time.Hour)),
	}

	// Call gRPC service
	response, err := h.urlGenClient.GenerateShortURL(ctx, grpcReq)
	if err != nil {
		createShortURLTotal.WithLabelValues("error").Inc()
		h.logger.Errorf("gRPC call to GenerateShortURL failed: %v", err)
		http.Error(w, "Failed to generate short URL", http.StatusInternalServerError)
		return
	}

	// Prepend domain to the short URL code
	response.ShortUrl = "http://" + URLShortenerDomainName + "/" + response.ShortUrl

	// Respond with the short URL
	responseJSON, err := json.Marshal(response)
	if err != nil {
		createShortURLTotal.WithLabelValues("error").Inc()
		h.logger.Errorf("Failed to marshal gRPC response: %v", err)
		http.Error(w, "Failed to process response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(responseJSON)
	if err != nil {
		createShortURLTotal.WithLabelValues("error").Inc()
		h.logger.Errorf("Failed to write response: %v", err)
		return
	}

	// Success - increment counter
	createShortURLTotal.WithLabelValues("success").Inc()
	h.logger.Infof("Successfully processed CreateShortURL request for %s", req.LongUrl)
}

func initLogger() *logrus.Logger {
	logger := logrus.New()
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		PadLevelText:  true,
	})
	logger.Infof("Logger initialized with log level: %s", logLevel)
	return logger
}
