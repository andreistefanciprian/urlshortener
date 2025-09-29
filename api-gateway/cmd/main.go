package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	ugen "github.com/andreistefanciprian/urlshortener/api-gateway/url-gen/proto"
	uread "github.com/andreistefanciprian/urlshortener/api-gateway/url-read/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {

	// Initialize the logger
	logger := initLogger()

	// Connect to URL Generation service
	urlGenHost := os.Getenv("URL_GEN_HOST")
	urlGenPort := os.Getenv("URL_GEN_PORT")
	urlGenAddress := fmt.Sprintf("%s:%s", urlGenHost, urlGenPort)
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
	urlReadHost := os.Getenv("URL_READ_HOST")
	urlReadPort := os.Getenv("URL_READ_PORT")
	urlReadAddress := fmt.Sprintf("%s:%s", urlReadHost, urlReadPort)
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
	apiHandler := NewAPIHandler(logger, urlGenClient, urlReadClient)
	httpPort := os.Getenv("API_GATEWAY_PORT")
	if httpPort == "" {
		httpPort = "8080" // Default port
	}
	// Set up HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /create", apiHandler.CreateShortURL)
	mux.HandleFunc("GET /{shortCode}", apiHandler.GetLongURL)

	// Start HTTP server
	logger.Infof("Starting API Gateway on port %s", httpPort)
	if err := http.ListenAndServe(":"+httpPort, mux); err != nil {
		logger.WithError(err).Fatal("Failed to start HTTP server")
	}

}

type HTTPHandler interface {
	// Define methods for handling HTTP requests
	CreateShortURL(w http.ResponseWriter, r *http.Request)
	GetLongURL(w http.ResponseWriter, r *http.Request)
}

type APIHandler struct {
	logger        *logrus.Logger
	urlGenClient  ugen.URLGeneratorClient // gRPC URLGeneratorClient
	urlReadClient uread.URLReaderClient   // gRPC URLReaderClient
}

func NewAPIHandler(logger *logrus.Logger, urlGenClient ugen.URLGeneratorClient, urlReadClient uread.URLReaderClient) *APIHandler {
	return &APIHandler{
		logger:        logger,
		urlGenClient:  urlGenClient,
		urlReadClient: urlReadClient,
	}
}

func (h *APIHandler) GetLongURL(w http.ResponseWriter, r *http.Request) {
	// Use request context for future enhancements such as logging, timeouts, tracing, etc.
	ctx := r.Context()

	// Extract short URL code from URL path
	shortUrlCode := r.PathValue("shortCode")

	// Validate required fields
	if err := validateShortURL(shortUrlCode); err != nil {
		h.logger.Errorf(err.Error())
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
		h.logger.Errorf("gRPC call to GetLongURL failed: %v", err)
		http.Error(w, "Failed to retrieve long URL", http.StatusInternalServerError)
		return
	}

	// Check if response.Expiration is set and if the URL has expired
	if response.Expiration != nil {
		if response.Expiration.AsTime().Before(time.Now()) {
			h.logger.Infof("Short URL %s has expired at %v", shortUrlCode, response.Expiration.AsTime())
			http.Error(w, "Short URL has expired", http.StatusGone)
			return
		}
	}

	// Redirect with 302 to the original long URL
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Expires", "0")
	http.Redirect(w, r, response.LongUrl, http.StatusFound)
}

const shortUrlCodeLength = 7

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

func validateCreateRequest(req *CreateShortURLRequest) error {
	// Validate required fields
	if req.LongUrl == "" {
		return fmt.Errorf("long URL is required")
	}
	// Set default expiration if not provided
	if req.ExpiresIn == 0 {
		req.ExpiresIn = 7 // Default to 7 days if not provided
	}
	// Further validation logic can be added here (e.g., regex check)
	return nil
}

func (h *APIHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	// Use request context for future enhancements such as logging, timeouts, tracing, etc.
	ctx := r.Context()

	// Parse request body
	var req CreateShortURLRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Errorf("Failed to decode JSON body: %v", err)
		http.Error(w, "Failed to decode JSON body", http.StatusBadRequest)
		return
	}

	// Validate required fields and set defaults
	if err := validateCreateRequest(&req); err != nil {
		h.logger.Errorf("Request validation failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create gRPC request
	grpcReq := &ugen.ShortURLRequest{
		LongUrl:    req.LongUrl,
		Expiration: timestamppb.New(time.Now().Add(time.Duration(req.ExpiresIn) * 24 * time.Hour)),
	}

	// Call gRPC service
	response, err := h.urlGenClient.GenerateShortURL(ctx, grpcReq)
	if err != nil {
		h.logger.Errorf("gRPC call to GenerateShortURL failed: %v", err)
		http.Error(w, "Failed to generate short URL", http.StatusInternalServerError)
		return
	}

	// Respond with the short URL
	responseJSON, err := json.Marshal(response)
	if err != nil {
		h.logger.Errorf("Failed to marshal gRPC response: %v", err)
		http.Error(w, "Failed to process response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(responseJSON)
	if err != nil {
		h.logger.Errorf("Failed to write response: %v", err)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
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
