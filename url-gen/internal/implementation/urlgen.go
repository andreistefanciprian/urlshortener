package urlgen

import (
	"context"
	"strings"

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

const (
	codeAlphabet = "FxnXM1kBN6cuhsAvjW3Co7l2RePyY8DwaU04Tzt9fHQrqSVKdpimLGIJOgb5ZE"
	codeLength   = 7 // 7-char length
)

func shortCodeGenerator() (string, error) {
	return gonanoid.Generate(codeAlphabet, codeLength)
}

func (s *URLGenService) GenerateShortURL(ctx context.Context, URLRequestPayload *ugen.ShortURLRequest) (*ugen.ShortURLResponse, error) {
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

	s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"shortURLCode": shortURLCode,
	}).Info("Successfully generated short URL code")

	return &ugen.ShortURLResponse{
		ShortUrl: shortURLCode,
	}, nil
}

func (s *URLGenService) DeleteShortURL(ctx context.Context, request *ugen.DeleteShortURLRequest) (*ugen.DeleteShortURLResponse, error) {
	err := s.repo.Delete(ctx, request.ShortUrlCode)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.logger.WithContext(ctx).WithFields(logrus.Fields{
				"shortURLCode": request.ShortUrlCode,
			}).Info("Short URL code not found in database")
			return &ugen.DeleteShortURLResponse{
				Success: false,
				Message: err.Error(),
			}, nil
		}
		s.logger.WithContext(ctx).WithFields(logrus.Fields{
			"shortURLCode": request.ShortUrlCode,
		}).WithError(err).Error("Failed to delete short URL Code")
		return &ugen.DeleteShortURLResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	// To be added Delete from Cache as well

	s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"shortURLCode": request.ShortUrlCode,
	}).Info("Successfully deleted short URL Code")

	return &ugen.DeleteShortURLResponse{
		Success: true,
		Message: "Short URL Code deleted successfully",
	}, nil
}
