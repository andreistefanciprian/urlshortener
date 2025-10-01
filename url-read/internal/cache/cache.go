package cache

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// URLCacher defines the interface for caching URL mappings
type URLCacher interface {
	Get(ctx context.Context, shortURL string) (string, error)
	Set(ctx context.Context, shortURL, longUrl string, expiresAt time.Time) error
	Del(ctx context.Context, shortURL string) error
}

// RedisURLCacher implements URLCacher interface for Redis cache
type RedisURLCacher struct {
	client *redis.Client
	logger *logrus.Logger
}

// NewRedisURLCacher creates a new Redis client for caching URL mappings
func NewRedisURLCacher(cache *redis.Client, logger *logrus.Logger) URLCacher {
	return &RedisURLCacher{
		client: cache,
		logger: logger,
	}
}

// Del deletes Short URL from the Redis cache
func (r *RedisURLCacher) Del(ctx context.Context, shortURL string) error {
	result, err := r.client.Del(ctx, shortURL).Result()
	if err != nil {
		r.logger.WithError(err).WithField("shortURL", shortURL).Error("Error deleting short URL from cache")
		return err
	}
	if result == 0 {
		r.logger.WithField("shortURL", shortURL).Info("Cache MISS: Short URL not found in Redis")
		return nil // No error for cache miss, just log it
	}
	// Successfully deleted from cache
	r.logger.WithField("shortURL", shortURL).Info("Short URL deleted from cache")
	return nil
}

// Get retrieves Long URL from the Redis cache by short URL
// Returns empty string and nil error if not found (cache miss)
func (r *RedisURLCacher) Get(ctx context.Context, shortURL string) (string, error) {
	longURL, err := r.client.Get(ctx, shortURL).Result()
	if err != nil {
		// Check if it's a cache miss (key not found)
		if err == redis.Nil {
			r.logger.WithField("shortURL", shortURL).Info("Cache MISS: Short URL not found in Redis, checking database")
			return "", nil // Return empty string and nil error for cache miss
		}
		// This is an actual Redis error (connection issues, etc.)
		r.logger.WithError(err).WithField("shortURL", shortURL).Error("Error retrieving long URL from cache")
		return "", err
	}

	r.logger.WithFields(logrus.Fields{
		"shortURL": shortURL,
		"longURL":  longURL,
	}).Info("Cache HIT: Short URL found in Redis")
	return longURL, nil
}

// Set stores Short URL and Long URL mapping in Redis cache with specified expiration time
func (r *RedisURLCacher) Set(ctx context.Context, shortURL, longURL string, expiresAt time.Time) error {
	// Calculate TTL from absolute expiration time
	// Negative and zero TTL validation will be done upstream
	ttl := time.Until(expiresAt)

	err := r.client.Set(ctx, shortURL, longURL, ttl).Err()
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"shortURL":  shortURL,
			"longURL":   longURL,
			"expiresAt": expiresAt,
		}).Error("Error storing long URL in cache")
		return err
	}
	r.logger.WithFields(logrus.Fields{
		"shortURL":  shortURL,
		"expiresAt": expiresAt,
		"ttl":       ttl,
	}).Info("Short URL stored in cache")
	return nil
}
