package cache

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/andreistefanciprian/urlshortener/url-read/internal/cache/testhelpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type URLCacheTestSuite struct {
	suite.Suite
	cacheContainer *testhelpers.RedisContainer
	cache          *RedisURLCache
	ctx            context.Context
}

func (c *URLCacheTestSuite) SetupSuite() {
	c.ctx = context.Background()
	cacheContainer, err := testhelpers.CreateRedisContainer(c.ctx)
	if err != nil {
		log.Fatal(err)
	}
	c.cacheContainer = cacheContainer
	logger := logrus.New()
	cache, err := NewRedisURLCache(logger, c.cacheContainer.RedisOpts)
	if err != nil {
		log.Fatal(err)
	}
	c.cache = cache
}

func (c *URLCacheTestSuite) TearDownSuite() {
	if err := c.cache.Close(); err != nil {
		log.Printf("Failed to close cache: %v", err)
	}
	if err := c.cacheContainer.RedisContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating redis container: %s", err)
	}
}

func (c *URLCacheTestSuite) TestSetAndGet_URLCacheEntry() {
	t := c.T()
	shortURLCode := "TEST123"
	expiresAt := time.Now().Add(10 * time.Minute)
	URLCacheEntry := URLCacheEntry{
		LongURL:   "https://example.com/long-url",
		ExpiresAt: expiresAt,
	}

	// Set the cached URL
	err := c.cache.Set(c.ctx, shortURLCode, URLCacheEntry)
	assert.NoError(t, err, "Failed to set cached URL")

	// Get the cached URL
	retrievedURL, err := c.cache.Get(c.ctx, shortURLCode)
	assert.NoError(t, err, "Failed to get cached URL")
	assert.NotNil(t, retrievedURL, "Expected to retrieve a cached URL, got nil")
	assert.Equal(t, URLCacheEntry.LongURL, retrievedURL.LongURL, "LongURL should match")
}

func (c *URLCacheTestSuite) TestGet_NonExistentKey() {
	t := c.T()
	shortURLCode := "nonexistent"

	retrievedURL, err := c.cache.Get(c.ctx, shortURLCode)
	assert.NoError(t, err, "Should not error when getting non-existent key")
	assert.Nil(t, retrievedURL, "Expected nil for non-existent key")
}

func TestURLCacheTestSuite(t *testing.T) {
	suite.Run(t, new(URLCacheTestSuite))
}
