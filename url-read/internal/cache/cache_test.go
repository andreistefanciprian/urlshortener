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
	redisContainer *testhelpers.RedisContainer
	cache          *RedisURLCache
	ctx            context.Context
}

func (suite *URLCacheTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	redisContainer, err := testhelpers.CreateRedisContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.redisContainer = redisContainer
	logger := logrus.New()
	cache, err := NewRedisURLCache(logger, suite.redisContainer.RedisOpts)
	if err != nil {
		log.Fatal(err)
	}
	suite.cache = cache
}

func (suite *URLCacheTestSuite) TearDownSuite() {
	if err := suite.cache.Close(); err != nil {
		log.Printf("Failed to close cache: %v", err)
	}
	if err := suite.redisContainer.RedisContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating redis container: %s", err)
	}
}

func (suite *URLCacheTestSuite) TestSetAndGet_URLCacheEntry() {
	t := suite.T()
	shortURLCode := "TEST123"
	expiresAt := time.Now().Add(10 * time.Minute)
	URLCacheEntry := URLCacheEntry{
		LongURL:   "https://example.com/long-url",
		ExpiresAt: expiresAt,
	}

	// Set the cached URL
	err := suite.cache.Set(suite.ctx, shortURLCode, URLCacheEntry)
	assert.NoError(t, err, "Failed to set cached URL")

	// Get the cached URL
	retrievedURL, err := suite.cache.Get(suite.ctx, shortURLCode)
	assert.NoError(t, err, "Failed to get cached URL")
	assert.NotNil(t, retrievedURL, "Expected to retrieve a cached URL, got nil")
	assert.Equal(t, URLCacheEntry.LongURL, retrievedURL.LongURL, "LongURL should match")
}

func (suite *URLCacheTestSuite) TestGet_NonExistentKey() {
	t := suite.T()
	shortURLCode := "nonexistent"

	retrievedURL, err := suite.cache.Get(suite.ctx, shortURLCode)
	if err != nil {
		t.Fatalf("Error occurred while getting non-existent key: %v", err)
	}

	assert.Nil(t, retrievedURL, "Expected nil for non-existent key")
}

func TestURLCacheTestSuite(t *testing.T) {
	suite.Run(t, new(URLCacheTestSuite))
}
