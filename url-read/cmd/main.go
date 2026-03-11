package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andreistefanciprian/urlshortener/url-read/internal/cache"
	"github.com/andreistefanciprian/urlshortener/url-read/internal/db"
	urlread "github.com/andreistefanciprian/urlshortener/url-read/internal/implementation"
	proto "github.com/andreistefanciprian/urlshortener/url-read/proto"
	redis "github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
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

// healthServer implements grpc_health_v1.HealthServer.
// Liveness (service="") always returns SERVING.
// Readiness (any named service) checks Postgres and Redis on each probe call.
type healthServer struct {
	grpc_health_v1.UnimplementedHealthServer
	logger   *logrus.Logger
	urlRepo  *db.PostgresURLStore
	urlCache *cache.RedisURLCache
}

func (h *healthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	// Liveness probe uses empty service name — always report serving
	if req.Service == "" {
		h.logger.Debug("Liveness probe hit")
		return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
	}

	// Readiness probe — check actual dependencies
	h.logger.Debugf("Readiness probe hit for service: %s", req.Service)
	checkCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	pgErr := h.urlRepo.HealthCheck(checkCtx)
	redisErr := h.urlCache.HealthCheck(checkCtx)

	if redisErr != nil {
		h.logger.Warnf("Redis unavailable, serving in degraded mode: %v", redisErr)
	}
	if pgErr != nil {
		h.logger.Warnf("Readiness check failed - postgres: %v", pgErr)
		return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING}, nil
	}

	h.logger.Debug("Readiness probe passed")
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
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
	urlReadService := urlread.NewURLReadService(logger, urlRepo, urlCache)
	proto.RegisterURLReaderServer(grpcServer, urlReadService)

	// Register gRPC health service (used by both liveness and readiness probes)
	grpc_health_v1.RegisterHealthServer(grpcServer, &healthServer{
		logger:   logger,
		urlRepo:  urlRepo,
		urlCache: urlCache,
	})
	logger.Info("gRPC health check service registered")
	if os.Getenv("GRPC_REFLECTION") == "true" {
		reflection.Register(grpcServer)
		logger.Info("gRPC reflection enabled")
	}

	// Set up context that cancels on SIGTERM/SIGINT for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// Start listening for gRPC requests
	serverPort := os.Getenv("URL_READ_PORT")
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", serverPort))
	if err != nil {
		logger.WithError(err).Fatal("Failed to start gRPC listener")
	}
	logger.Infof("gRPC server listening at %v", listener.Addr())

	// Shut down gRPC server gracefully when context is cancelled
	go func() {
		<-ctx.Done()
		logger.Info("Shutting down gRPC server")
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(listener); err != nil && ctx.Err() == nil {
		logger.WithError(err).Error("gRPC server exited unexpectedly")
	}
	logger.Info("gRPC server stopped")
}
