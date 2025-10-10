package urlgen

import (
	"context"
	"log"
	"testing"

	cache "github.com/andreistefanciprian/urlshortener/url-gen/internal/cache"
	db "github.com/andreistefanciprian/urlshortener/url-gen/internal/db"
	"github.com/andreistefanciprian/urlshortener/url-gen/internal/implementation/testhelpers"
	ugen "github.com/andreistefanciprian/urlshortener/url-gen/proto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/emptypb"
)

type URLGenServiceTestSuite struct {
	suite.Suite
	dbContainer    *testhelpers.PostgresContainer
	cacheContainer *testhelpers.RedisContainer
	repository     *db.PostgresURLStore
	cache          *cache.RedisURLCache
	implementation *URLGenService
	ctx            context.Context
}

func (c *URLGenServiceTestSuite) SetupSuite() {
	c.ctx = context.Background()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

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

	// Create the URLGenService instance with real dependencies
	c.implementation = NewURLGenService(logger, c.repository, c.cache)
}

func (c *URLGenServiceTestSuite) TestGenerateShortURL_Length() {
	t := c.T()

	response, err := c.implementation.GenerateShortURL(c.ctx, &ugen.ShortURLRequest{LongUrl: "https://example.com/long-url-1"})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 7, len(response.ShortUrl))
}

func (c *URLGenServiceTestSuite) TestDeleteShortURL_Success() {
	t := c.T()

	// First, generate a short URL
	response, err := c.implementation.GenerateShortURL(c.ctx, &ugen.ShortURLRequest{LongUrl: "https://example.com/long-url-1"})
	assert.NoError(t, err)
	assert.NotNil(t, response)

	// Now, delete the short URL
	deleteResponse, err := c.implementation.DeleteShortURL(c.ctx, &ugen.DeleteShortURLRequest{ShortUrlCode: response.ShortUrl})
	assert.NoError(t, err)
	assert.NotNil(t, deleteResponse)
	assert.True(t, deleteResponse.Success)
}

func (c *URLGenServiceTestSuite) TestDeleteShortURL_DoesNotExist() {
	t := c.T()

	// Attempt to delete a non-existent short URL
	deleteResponse, err := c.implementation.DeleteShortURL(c.ctx, &ugen.DeleteShortURLRequest{ShortUrlCode: "NONEXIST"})
	assert.NoError(t, err)
	assert.NotNil(t, deleteResponse)
	assert.False(t, deleteResponse.Success)
}

func (c *URLGenServiceTestSuite) TestGetAllURLs_Success() {
	t := c.T()

	response, err := c.implementation.GetAllURLs(c.ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Urls)
	assert.Equal(t, 2, len(response.Urls)) // Only 2 non-expired URLs in test data
}

func (c *URLGenServiceTestSuite) TestGetAllURLs_AfterDeletinngAllURLs() {
	t := c.T()

	// First, delete all existing URLs
	response, err := c.implementation.GetAllURLs(c.ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, response)

	for _, url := range response.Urls {
		_, err := c.implementation.DeleteShortURL(c.ctx, &ugen.DeleteShortURLRequest{ShortUrlCode: url.ShortUrlCode})
		assert.NoError(t, err)
	}

	// Now, get all URLs again
	responseAfterDeletion, err := c.implementation.GetAllURLs(c.ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, responseAfterDeletion)
	assert.Empty(t, responseAfterDeletion.Urls)
	assert.Equal(t, 0, len(responseAfterDeletion.Urls))
}

func (c *URLGenServiceTestSuite) TearDownSuite() {
	if err := c.repository.Close(); err != nil {
		log.Printf("Failed to close repository: %v", err)
	}
	if err := c.dbContainer.PostgresContainer.Terminate(c.ctx); err != nil {
		log.Printf("error terminating postgres container: %s", err)
	}

	if err := c.cache.Close(); err != nil {
		log.Printf("Failed to close cache: %v", err)
	}
	if err := c.cacheContainer.RedisContainer.Terminate(c.ctx); err != nil {
		log.Printf("error terminating redis container: %s", err)
	}
}

func TestURLGenServiceTestSuite(t *testing.T) {
	suite.Run(t, new(URLGenServiceTestSuite))
}
