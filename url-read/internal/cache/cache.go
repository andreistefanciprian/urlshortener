package cache

import (
	"context"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// URLCacheEntry represents a cached URL with its expiration time
type URLCacheEntry struct {
	LongURL   string    `json:"longUrl"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// URLCache defines the interface for caching URL mappings
type URLCache interface {
	Get(ctx context.Context, shortURLCode string) (*URLCacheEntry, error)
	Set(ctx context.Context, shortURLCode string, cacheEntry URLCacheEntry) error
	Del(ctx context.Context, shortURLCode string) error
	Close() error
}

// RedisURLCache implements URLCache interface for Redis cache
type RedisURLCache struct {
	client *redis.Client
	logger *logrus.Logger
}

// NewRedisURLCache creates a new Redis client for caching URL mappings
func NewRedisURLCache(logger *logrus.Logger, redisOptions *redis.Options) (*RedisURLCache, error) {
	// Initialize Redis client
	cache, err := initCache(redisOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis cache: %w", err)
	}

	return &RedisURLCache{
		client: cache,
		logger: logger,
	}, nil
}

// Del deletes Short URL Code from the Redis cache
func (r *RedisURLCache) Del(ctx context.Context, shortURLCode string) error {
	result, err := r.client.Del(ctx, shortURLCode).Result()
	if err != nil {
		r.logger.WithError(err).WithField("shortURLCode", shortURLCode).Error("Error deleting short URL Code from cache")
		return err
	}
	if result == 0 {
		r.logger.WithField("shortURLCode", shortURLCode).Info("Cache MISS: Short URL Code not found in cache")
		return nil // No error for cache miss, just log it
	}
	// Successfully deleted from cache
	r.logger.WithField("shortURLCode", shortURLCode).Info("Short URL Code deleted from cache")
	return nil
}

// Get retrieves Long URL from the Redis cache by short URL Code
// Returns URLCacheEntry struct with longURL and calculated expiration time
func (r *RedisURLCache) Get(ctx context.Context, shortURLCode string) (*URLCacheEntry, error) {
	// Get the long URL
	longURL, err := r.client.Get(ctx, shortURLCode).Result()
	if err != nil {
		// Check if it's a cache miss (key not found)
		if err == redis.Nil {
			r.logger.WithField("shortURLCode", shortURLCode).Info("Cache MISS: Short URL Code not found in cache")
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

	cacheEntry := &URLCacheEntry{
		LongURL:   longURL,
		ExpiresAt: expiresAt,
	}

	r.logger.WithFields(logrus.Fields{
		"shortURLCode": shortURLCode,
		"longURL":      longURL,
		"expiresAt":    expiresAt,
		"ttl":          ttl,
	}).Info("Cache HIT: Short URL Code found in cache")

	return cacheEntry, nil
}

// Set stores Short URL Code and Long URL mapping in Redis cache with specified expiration time
func (r *RedisURLCache) Set(ctx context.Context, shortURLCode string, cacheEntry URLCacheEntry) error {
	// Calculate TTL from absolute expiration time
	// Negative and zero TTL validation will be done upstream
	ttl := time.Until(cacheEntry.ExpiresAt)

	err := r.client.Set(ctx, shortURLCode, cacheEntry.LongURL, ttl).Err()
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"shortURLCode": shortURLCode,
			"longURL":      cacheEntry.LongURL,
			"expiresAt":    cacheEntry.ExpiresAt,
		}).Error("Error storing long URL in cache")
		return err
	}
	r.logger.WithFields(logrus.Fields{
		"shortURLCode": shortURLCode,
		"expiresAt":    cacheEntry.ExpiresAt,
		"ttl":          ttl,
	}).Info("Short URL Code stored in cache")
	return nil
}

// Close closes the Redis client connection
func (r *RedisURLCache) Close() error {
	return r.client.Close()
}

// HealthCheck verifies the Redis connection is alive
func (r *RedisURLCache) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
