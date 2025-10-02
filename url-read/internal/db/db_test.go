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
	pgContainer *testhelpers.PostgresContainer
	repository  *MyURLRepository
	ctx         context.Context
}

func (suite *URLRepoTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	pgContainer, err := testhelpers.CreatePostgresContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.pgContainer = pgContainer

	logger := logrus.New()
	repository, err := NewMyURLRepository(logger, suite.pgContainer.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	suite.repository = repository
}

func (suite *URLRepoTestSuite) TearDownSuite() {
	if err := suite.repository.Close(); err != nil {
		log.Printf("Failed to close repository: %v", err)
	}
	if err := suite.pgContainer.PostgresContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
}

func (suite *URLRepoTestSuite) TestGetLongURL_ValidURL() {
	t := suite.T()

	longURLResponse, err := suite.repository.GetLongURL(suite.ctx, "wRZ32pT")
	assert.NoError(t, err)
	assert.NotNil(t, longURLResponse)
	assert.Equal(t, "https://example.com/long-url-1", longURLResponse.OriginalURL)
	assert.NotNil(t, longURLResponse.ExpiresAt)
}

func (suite *URLRepoTestSuite) TestGetLongURL_NonExistentURL() {
	t := suite.T()

	_, err := suite.repository.GetLongURL(suite.ctx, "INVALID")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func (suite *URLRepoTestSuite) TestGetLongURL_ExpiredURL() {
	t := suite.T()

	// N9qGJxH is an expired URL in the test data
	_, err := suite.repository.GetLongURL(suite.ctx, "N9qGJxH")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func (suite *URLRepoTestSuite) TestGetLongURL_URLWithoutExpiration() {
	t := suite.T()

	// GD38yB5 has no expiration (NULL expires_at) in the test data
	longURLResponse, err := suite.repository.GetLongURL(suite.ctx, "GD38yB5")
	assert.NoError(t, err)
	assert.NotNil(t, longURLResponse)
	assert.Equal(t, "https://example.com/long-url-2", longURLResponse.OriginalURL)
	assert.Nil(t, longURLResponse.ExpiresAt)
}

func TestURLRepoTestSuite(t *testing.T) {
	suite.Run(t, new(URLRepoTestSuite))
}
