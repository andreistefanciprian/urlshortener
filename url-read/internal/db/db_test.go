package db

import (
	"context"
	"log"
	"testing"

	"github.com/andreistefanciprian/urlshortener/url-read/internal/db/testhelpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type URLRepoTestSuite struct {
	suite.Suite
	dbContainer *testhelpers.PostgresContainer
	repository  *PostgresURLStore
	ctx         context.Context
}

func (c *URLRepoTestSuite) SetupSuite() {
	c.ctx = context.Background()
	dbContainer, err := testhelpers.CreatePostgresContainer(c.ctx)
	if err != nil {
		log.Fatal(err)
	}
	c.dbContainer = dbContainer

	logger := logrus.New()
	repository, err := NewPostgresURLStore(logger, c.dbContainer.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	c.repository = repository
}

func (c *URLRepoTestSuite) TearDownSuite() {
	if err := c.repository.Close(); err != nil {
		log.Printf("Failed to close repository: %v", err)
	}
	if err := c.dbContainer.PostgresContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
}

func (c *URLRepoTestSuite) TestGetLongURL_ValidURL() {
	t := c.T()

	longURLResponse, err := c.repository.GetLongURL(c.ctx, "wRZ32pT")
	assert.NoError(t, err)
	assert.NotNil(t, longURLResponse)
	assert.Equal(t, "https://example.com/long-url-1", longURLResponse.OriginalURL)
	assert.NotNil(t, longURLResponse.ExpiresAt)
}

func (c *URLRepoTestSuite) TestGetLongURL_NonExistentURL() {
	t := c.T()

	_, err := c.repository.GetLongURL(c.ctx, "INVALID")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func (c *URLRepoTestSuite) TestGetLongURL_ExpiredURL() {
	t := c.T()

	// N9qGJxH is an expired URL in the test data
	_, err := c.repository.GetLongURL(c.ctx, "N9qGJxH")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func (c *URLRepoTestSuite) TestGetLongURL_URLWithoutExpiration() {
	t := c.T()

	// GD38yB5 has no expiration (NULL expires_at) in the test data
	longURLResponse, err := c.repository.GetLongURL(c.ctx, "GD38yB5")
	assert.NoError(t, err)
	assert.NotNil(t, longURLResponse)
	assert.Equal(t, "https://example.com/long-url-2", longURLResponse.OriginalURL)
	assert.Nil(t, longURLResponse.ExpiresAt)
}

func TestURLRepoTestSuite(t *testing.T) {
	suite.Run(t, new(URLRepoTestSuite))
}
