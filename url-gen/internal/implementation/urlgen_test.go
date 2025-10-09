package urlgen

import (
	"context"
	"log"
	"testing"

	db "github.com/andreistefanciprian/urlshortener/url-gen/internal/db"
	"github.com/andreistefanciprian/urlshortener/url-gen/internal/implementation/testhelpers"
	ugen "github.com/andreistefanciprian/urlshortener/url-gen/proto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type URLGenServiceTestSuite struct {
	suite.Suite
	dbContainer    *testhelpers.PostgresContainer
	repository     *db.PostgresURLStore
	implementation *URLGenService
	ctx            context.Context
}

func (c *URLGenServiceTestSuite) SetupSuite() {
	c.ctx = context.Background()
	logger := logrus.New()

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

	// Create the URLGenService instance with real dependencies
	c.implementation = NewURLGenService(logger, c.repository)
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

func (c *URLGenServiceTestSuite) TearDownSuite() {
	if err := c.repository.Close(); err != nil {
		log.Printf("Failed to close repository: %v", err)
	}
	if err := c.dbContainer.PostgresContainer.Terminate(c.ctx); err != nil {
		log.Printf("error terminating postgres container: %s", err)
	}
}

func TestURLGenServiceTestSuite(t *testing.T) {
	suite.Run(t, new(URLGenServiceTestSuite))
}
