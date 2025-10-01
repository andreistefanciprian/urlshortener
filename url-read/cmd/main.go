package main

import (
	"fmt"
	"net"
	"os"

	"github.com/andreistefanciprian/urlshortener/url-read/internal/cache"
	"github.com/andreistefanciprian/urlshortener/url-read/internal/db"
	urlread "github.com/andreistefanciprian/urlshortener/url-read/internal/implementation"
	proto "github.com/andreistefanciprian/urlshortener/url-read/proto"
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
	logger.Info("Connected to PostgreSQL database")

	// Initialize Redis URL Cacher
	urlCacher, err := cache.NewRedisURLCacher(logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Redis URL cacher")
	}
	defer urlCacher.Close()
	logger.Info("Connected to Redis cache")

	// Create gRPC server
	grpcServer := grpc.NewServer()
	urlReadService := urlread.NewUrlReadImplementation(logger, repo, urlCacher)
	proto.RegisterURLReaderServer(grpcServer, urlReadService)

	// Start listening for gRPC requests
	serverPort := os.Getenv("URL_READ_PORT")
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", serverPort))
	if err != nil {
		logger.WithError(err).Fatal("Failed to start gRPC listener")
	}
	logger.Infof("gRPC server listening at %v", listener.Addr())
	if err := grpcServer.Serve(listener); err != nil {
		logger.WithError(err).Fatal("Failed to serve gRPC server")
	}
}
