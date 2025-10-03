package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/andreistefanciprian/urlshortener/url-read/internal/cache"
	"github.com/andreistefanciprian/urlshortener/url-read/internal/db"
	urlread "github.com/andreistefanciprian/urlshortener/url-read/internal/implementation"
	proto "github.com/andreistefanciprian/urlshortener/url-read/proto"
	redis "github.com/redis/go-redis/v9"
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
	dbConfig := db.GetDBConfigFromEnv()
	urlRepo, err := db.NewPostgresURLStore(logger, dbConfig.DSNString())
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize database repository")
	}
	defer urlRepo.Close()
	logger.Info("Connected to PostgreSQL database")

	// Initialize Redis URL Cacher
	cacheConfig := cache.GetCacheConfigFromEnv()
	redisOptions := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cacheConfig.Host, cacheConfig.Port),
		Password: cacheConfig.Password,
		DB:       cacheConfig.DB,
		// Connection pool settings optimized for read operations
		PoolSize:        20,              // Max connections in pool
		MinIdleConns:    5,               // Keep connections warm
		PoolTimeout:     time.Second * 4, // Wait time for pool connection
		ConnMaxIdleTime: time.Minute * 5, // Close idle connections after 5 minutes
	}

	urlCache, err := cache.NewRedisURLCache(logger, redisOptions)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Redis URL cacher")
	}
	defer urlCache.Close()
	logger.Info("Connected to Redis cache")

	// Create gRPC server
	grpcServer := grpc.NewServer()
	urlReadService := urlread.NewUrlReadImplementation(logger, urlRepo, urlCache)
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
