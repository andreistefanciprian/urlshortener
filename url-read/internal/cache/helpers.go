package cache

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// CacheConfig holds Redis cache configuration parameters
type CacheConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// initCache collects cache parameters from environment variables and creates Redis connection
func initCache(redisOptions *redis.Options) (*redis.Client, error) {
	// Suppress the Redis client's internal connection pool logs
	redis.SetLogger(log.New(io.Discard, "", 0))

	// Create Redis client
	rdb := redis.NewClient(redisOptions)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}
	return rdb, nil
}

// getCacheConfigFromEnv collects cache configuration from environment variables
func GetCacheConfigFromEnv() CacheConfig {
	config := CacheConfig{
		Host:     getEnvOrDefault("REDIS_HOST", "localhost"),
		Port:     getEnvAsIntOrDefault("REDIS_PORT", 6379),
		Password: getEnvOrDefault("REDIS_PASSWORD", "redispassword"),
		DB:       getEnvAsIntOrDefault("REDIS_DB", 0),
	}
	return config
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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
