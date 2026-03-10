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

var ErrShortURLCodeNotFound = fmt.Errorf("short URL code not found")

// Del deletes Short URL Code from the Redis cache
func (r *RedisURLCache) Del(ctx context.Context, shortURLCode string) error {
	result, err := r.client.Del(ctx, shortURLCode).Result()
	if err != nil {
		r.logger.WithError(err).WithField("shortURLCode", shortURLCode).Error("Error deleting short URL Code from cache")
		return err
	}
	if result == 0 {
		r.logger.WithFields(logrus.Fields{
			"shortURLCode": shortURLCode,
		}).Debug("Cache MISS: Short URL code not found in cache")
		return ErrShortURLCodeNotFound
	}
	// Successfully deleted from cache
	r.logger.WithField("shortURLCode", shortURLCode).Debug("Short URL Code deleted from cache")
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
