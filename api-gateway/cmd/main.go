package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	ugen "github.com/andreistefanciprian/urlshortener/url-gen/proto"
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

	// Initialize HTTP handler
	apiHandler := NewAPIHandler(logger, urlGenClient)
	httpPort := os.Getenv("API_GATEWAY_PORT")
	if httpPort == "" {
		httpPort = "8080" // Default port
	}
	// Set up HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /create", apiHandler.CreateShortURL)

	// Start HTTP server
	logger.Infof("Starting API Gateway on port %s", httpPort)
	if err := http.ListenAndServe(":"+httpPort, mux); err != nil {
		logger.WithError(err).Fatal("Failed to start HTTP server")
	}

}

type HTTPHandler interface {
	// Define methods for handling HTTP requests
	CreateShortURL(w http.ResponseWriter, r *http.Request)
}

type APIHandler struct {
	logger       *logrus.Logger
	urlGenClient ugen.URLGeneratorClient // gRPC URLGeneratorClient
}

func NewAPIHandler(logger *logrus.Logger, urlGenClient ugen.URLGeneratorClient) *APIHandler {
	return &APIHandler{
		logger:       logger,
		urlGenClient: urlGenClient, //
	}
}

func (h *APIHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	// Use request context for future enhancements such as logging, timeouts, tracing, etc.
	ctx := r.Context()

	// Parse request body
	var req struct {
		LongUrl   string `json:"longUrl"`
		ExpiresIn int    `json:"expiresIn"` // Expiration time in days
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Errorf("Failed to decode JSON body: %v", err)
		http.Error(w, "Failed to decode JSON body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.LongUrl == "" {
		h.logger.Error("Long URL is blank")
		http.Error(w, "Long URL is required", http.StatusBadRequest)
		return
	}
	if req.ExpiresIn == 0 {
		req.ExpiresIn = 7 // Default to 7 days if not provided
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
