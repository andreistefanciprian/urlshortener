package urlgen

import (
	"context"

	repo "github.com/andreistefanciprian/urlshortener/url-gen/internal/db"
	ugen "github.com/andreistefanciprian/urlshortener/url-gen/proto"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/sirupsen/logrus"
)

type URLGenService struct {
	logger *logrus.Logger
	repo   repo.URLStore
	ugen.UnimplementedURLGeneratorServer
}

func NewURLGenService(logger *logrus.Logger, repository repo.URLStore) *URLGenService {
	return &URLGenService{
		logger: logger,
		repo:   repository,
	}
}

func (s *URLGenService) GenerateShortURL(ctx context.Context, URLRequestPayload *ugen.ShortURLRequest) (*ugen.ShortURLResponse, error) {
	// Use context for logging and timeout handling
	s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"longUrl": URLRequestPayload.LongUrl,
	}).Info("Processing short URL generation request")

	// Check if context is cancelled
	if ctx.Err() != nil {
		s.logger.WithContext(ctx).Error("Request context cancelled")
		return nil, ctx.Err()
	}

	// Generate a unique short code
	shortURLCode, err := shortCodeGenerator()
	if err != nil {
		s.logger.WithContext(ctx).WithError(err).Error("Failed to generate short code")
		return nil, err
	}

	// Store the mapping in the database (context will be used by db layer)
	err = s.repo.CreateShortURL(ctx, URLRequestPayload.LongUrl, shortURLCode, URLRequestPayload.Expiration.AsTime())
	if err != nil {
		s.logger.WithContext(ctx).WithError(err).Error("Failed to create short URL")
		return nil, err
	}

	// Return the full short URL
	shortUrl := "http://" + URLShortenerDomainName + "/" + shortURLCode

	s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"shortURLCode": shortURLCode,
		"shortUrl":     shortUrl,
	}).Info("Successfully generated short URL")

	return &ugen.ShortURLResponse{
		ShortUrl: shortUrl,
	}, nil
}

const (
	codeAlphabet           = "FxnXM1kBN6cuhsAvjW3Co7l2RePyY8DwaU04Tzt9fHQrqSVKdpimLGIJOgb5ZE"
	codeLength             = 7 // 7-char length
	URLShortenerDomainName = "l.it"
)

func shortCodeGenerator() (string, error) {
	return gonanoid.Generate(codeAlphabet, codeLength)
}
