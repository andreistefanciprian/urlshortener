package urlread

import (
	"context"
	"log"
	"testing"
	"time"

	cache "github.com/andreistefanciprian/urlshortener/url-read/internal/cache"
	db "github.com/andreistefanciprian/urlshortener/url-read/internal/db"
	"github.com/andreistefanciprian/urlshortener/url-read/internal/implementation/testhelpers"
	uread "github.com/andreistefanciprian/urlshortener/url-read/proto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type URLReadServiceTestSuite struct {
	suite.Suite
	cacheContainer *testhelpers.RedisContainer
	dbContainer    *testhelpers.PostgresContainer
	cache          *cache.RedisURLCache
	repository     *db.PostgresURLStore
	implementation *URLReadService
	ctx            context.Context
}

func (c *URLReadServiceTestSuite) SetupSuite() {
	c.ctx = context.Background()
	logger := logrus.New()

	// Setup Cache container
	cacheContainer, err := testhelpers.CreateRedisContainer(c.ctx)
	if err != nil {
		log.Fatal(err)
	}
	c.cacheContainer = cacheContainer

	cache, err := cache.NewRedisURLCache(logger, c.cacheContainer.RedisOpts)
	if err != nil {
		log.Fatal(err)
	}
	c.cache = cache

	// Setup DB container
	dbContainer, err := testhelpers.CreatePostgresContainer(c.ctx)
	if err != nil {
		log.Fatal(err)
	}
	c.dbContainer = dbContainer

	repository, err := db.NewPostgresURLStore(logger, c.dbContainer.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	c.repository = repository

	// Create the URLReadService instance with real dependencies
	c.implementation = NewURLReadService(logger, c.repository, c.cache)
}

func (c *URLReadServiceTestSuite) TestGetLongURL_ValidURL() {
	t := c.T()

	response, err := c.implementation.GetLongURL(c.ctx, &uread.LongURLRequest{ShortUrl: "wRZ32pT"})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "https://example.com/long-url-1", response.LongUrl)
	assert.NotNil(t, response.Expiration)
}

func (c *URLReadServiceTestSuite) TestGetLongURL_NonExistentURL() {
	t := c.T()

	_, err := c.implementation.GetLongURL(c.ctx, &uread.LongURLRequest{ShortUrl: "INVALID"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func (c *URLReadServiceTestSuite) TestGetLongURL_ExpiredURL() {
	t := c.T()

	// N9qGJxH is an expired URL in the test data
	_, err := c.implementation.GetLongURL(c.ctx, &uread.LongURLRequest{ShortUrl: "N9qGJxH"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func (c *URLReadServiceTestSuite) TestGetLongURL_URLWithoutExpiration() {
	t := c.T()

	// GD38yB5 has no expiration (NULL expires_at) in the test data
	response, err := c.implementation.GetLongURL(c.ctx, &uread.LongURLRequest{ShortUrl: "GD38yB5"})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "https://example.com/long-url-2", response.LongUrl)
	assert.Nil(t, response.Expiration)
}

func (c *URLReadServiceTestSuite) TestGetLongUR_CacheHIT() {
	t := c.T()
	shortURLCode := "TEST123"
	expiresAt := time.Now().Add(10 * time.Minute)
	URLCacheEntry := cache.URLCacheEntry{
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

func (c *URLReadServiceTestSuite) TearDownSuite() {
	if err := c.cache.Close(); err != nil {
		log.Printf("Failed to close cache: %v", err)
	}
	if err := c.cacheContainer.RedisContainer.Terminate(c.ctx); err != nil {
		log.Printf("error terminating redis container: %s", err)
	}

	if err := c.repository.Close(); err != nil {
		log.Printf("Failed to close repository: %v", err)
	}
	if err := c.dbContainer.PostgresContainer.Terminate(c.ctx); err != nil {
		log.Printf("error terminating postgres container: %s", err)
	}
}

func TestURLReadServiceTestSuite(t *testing.T) {
	suite.Run(t, new(URLReadServiceTestSuite))
}
