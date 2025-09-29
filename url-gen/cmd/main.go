package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/andreistefanciprian/urlshortener/url-gen/internal/db"
	urlgen "github.com/andreistefanciprian/urlshortener/url-gen/internal/implementation"
	proto "github.com/andreistefanciprian/urlshortener/url-gen/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var logger = logrus.New()

func initLogger() {
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
}

func main() {
	// Initialize the logger
	initLogger()

	// Initialize DB repository
	repo, err := db.NewMyURLRepository(logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize database repository")
	}
	defer repo.Close()

	// Create gRPC server
	grpcServer := grpc.NewServer()
	urlGenService := urlgen.NewUrlGenImplementation(logger, repo)
	proto.RegisterURLGeneratorServer(grpcServer, urlGenService)

	// Start listening for gRPC requests
	serverPort := os.Getenv("URL_GEN_PORT")
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", serverPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("gRPC server listening at %v", listener.Addr())
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
