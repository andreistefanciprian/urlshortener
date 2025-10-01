package cache

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// CachedURL represents a cached URL with its expiration time
type CachedURL struct {
	LongURL   string    `json:"longUrl"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// URLCacher defines the interface for caching URL mappings
type URLCacher interface {
	Get(ctx context.Context, shortURLCode string) (*CachedURL, error)
	Set(ctx context.Context, shortURLCode string, cachedURL CachedURL) error
	Del(ctx context.Context, shortURLCode string) error
	Close() error
}

// RedisURLCacher implements URLCacher interface for Redis cache
type RedisURLCacher struct {
	client *redis.Client
	logger *logrus.Logger
}

// NewRedisURLCacher creates a new Redis client for caching URL mappings
func NewRedisURLCacher(logger *logrus.Logger) (URLCacher, error) {
	// Initialize Redis client
	cache, err := initCache()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis cache: %w", err)
	}

	return &RedisURLCacher{
		client: cache,
		logger: logger,
	}, nil
}

// Del deletes Short URL Code from the Redis cache
func (r *RedisURLCacher) Del(ctx context.Context, shortURLCode string) error {
	result, err := r.client.Del(ctx, shortURLCode).Result()
	if err != nil {
		r.logger.WithError(err).WithField("shortURLCode", shortURLCode).Error("Error deleting short URL Codefrom cache")
		return err
	}
	if result == 0 {
		r.logger.WithField("shortURLCode", shortURLCode).Info("Cache MISS: Short URL Codenot found in Redis")
		return nil // No error for cache miss, just log it
	}
	// Successfully deleted from cache
	r.logger.WithField("shortURLCode", shortURLCode).Info("Short URL Codedeleted from cache")
	return nil
}

// Get retrieves Long URL from the Redis cache by short URL Code
// Returns CachedURL struct with longURL and calculated expiration time
func (r *RedisURLCacher) Get(ctx context.Context, shortURLCode string) (*CachedURL, error) {
	// Get the long URL
	longURL, err := r.client.Get(ctx, shortURLCode).Result()
	if err != nil {
		// Check if it's a cache miss (key not found)
		if err == redis.Nil {
			r.logger.WithField("shortURLCode", shortURLCode).Info("Cache MISS: Short URL Code not found in Redis, checking database")
			return nil, nil // Return nil for cache miss
		}
		// This is an actual Redis error (connection issues, etc.)
		r.logger.WithError(err).WithField("shortURLCode", shortURLCode).Error("Error retrieving long URL from cache")
		return nil, err
	}

	// Get the remaining TTL and calculate expiration time
	// We calculate expiration time based on current time + TTL
	// faster than JSON unmarshalling if we had stored a struct
	ttl, err := r.client.TTL(ctx, shortURLCode).Result()
	if err != nil {
		r.logger.WithError(err).WithField("shortURLCode", shortURLCode).Error("Error retrieving TTL from cache")
		return nil, err
	}

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}
	// If TTL is -1 (no expiration) or -2 (key doesn't exist), expiresAt remains zero

	cachedURL := &CachedURL{
		LongURL:   longURL,
		ExpiresAt: expiresAt,
	}

	r.logger.WithFields(logrus.Fields{
		"shortURLCode": shortURLCode,
		"longURL":      longURL,
		"expiresAt":    expiresAt,
		"ttl":          ttl,
	}).Info("Cache HIT: Short URL Code found in Redis")

	return cachedURL, nil
}

// Set stores Short URL Code and Long URL mapping in Redis cache with specified expiration time
func (r *RedisURLCacher) Set(ctx context.Context, shortURLCode string, cachedURL CachedURL) error {
	// Calculate TTL from absolute expiration time
	// Negative and zero TTL validation will be done upstream
	ttl := time.Until(cachedURL.ExpiresAt)

	err := r.client.Set(ctx, shortURLCode, cachedURL.LongURL, ttl).Err()
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"shortURLCode": shortURLCode,
			"longURL":      cachedURL.LongURL,
			"expiresAt":    cachedURL.ExpiresAt,
		}).Error("Error storing long URL in cache")
		return err
	}
	r.logger.WithFields(logrus.Fields{
		"shortURLCode": shortURLCode,
		"expiresAt":    cachedURL.ExpiresAt,
		"ttl":          ttl,
	}).Info("Short URL Code stored in cache")
	return nil
}

// Close closes the Redis client connection
func (r *RedisURLCacher) Close() error {
	return r.client.Close()
}

// CacheConfig holds Redis cache configuration parameters
type CacheConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// initCache collects cache parameters from environment variables and creates Redis connection
func initCache() (*redis.Client, error) {
	// Get cache configuration from environment variables
	config := getCacheConfigFromEnv()

	// Create Redis client options
	opt := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,

		// Connection pool settings optimized for read operations
		PoolSize:        20,              // Max connections in pool
		MinIdleConns:    5,               // Keep connections warm
		PoolTimeout:     time.Second * 4, // Wait time for pool connection
		ConnMaxIdleTime: time.Minute * 5, // Close idle connections after 5 minutes
	}

	// Create Redis client
	rdb := redis.NewClient(opt)

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
func getCacheConfigFromEnv() CacheConfig {
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
