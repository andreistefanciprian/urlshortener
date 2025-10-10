package urlgen

import (
	"context"
	"errors"

	"github.com/andreistefanciprian/urlshortener/url-gen/internal/cache"
	repo "github.com/andreistefanciprian/urlshortener/url-gen/internal/db"
	ugen "github.com/andreistefanciprian/urlshortener/url-gen/proto"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type URLGenService struct {
	logger *logrus.Logger
	repo   repo.URLStore
	cache  cache.URLCache
	ugen.UnimplementedURLGeneratorServer
}

func NewURLGenService(logger *logrus.Logger, repository repo.URLStore, cache cache.URLCache) *URLGenService {
	return &URLGenService{
		logger: logger,
		repo:   repository,
		cache:  cache,
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

// TODO: maybe we should only return error and not a response
func (s *URLGenService) DeleteShortURL(ctx context.Context, request *ugen.DeleteShortURLRequest) (*ugen.DeleteShortURLResponse, error) {
	// Delete from Cache first
	err := s.cache.Del(ctx, request.ShortUrlCode)
	if err != nil {
		if errors.Is(err, cache.ErrShortURLCodeNotFound) {
			s.logger.WithContext(ctx).WithFields(logrus.Fields{
				"shortURLCode": request.ShortUrlCode,
			}).Info("Short URL code not found in cache")
		} else {
			s.logger.WithContext(ctx).WithFields(logrus.Fields{
				"shortURLCode": request.ShortUrlCode,
			}).WithError(err).Error("Failed to delete short URL code from cache")
		}
	} else {
		s.logger.WithContext(ctx).WithFields(logrus.Fields{
			"shortURLCode": request.ShortUrlCode,
		}).Info("Successfully deleted short URL Code from cache")
	}

	// Delete from Database
	err = s.repo.Delete(ctx, request.ShortUrlCode)
	if err != nil {
		if errors.Is(err, repo.ErrShortURLCodeNotFound) {
			s.logger.WithContext(ctx).WithFields(logrus.Fields{
				"shortURLCode": request.ShortUrlCode,
			}).Info("Short URL code not found in database")
			return &ugen.DeleteShortURLResponse{
				Success: false,
				Message: repo.ErrShortURLCodeNotFound.Error(),
			}, nil
		}
		s.logger.WithContext(ctx).WithFields(logrus.Fields{
			"shortURLCode": request.ShortUrlCode,
		}).WithError(err).Error("Failed to delete short URL Code")
		return &ugen.DeleteShortURLResponse{
			Success: false,
			Message: err.Error(),
		}, err
	} else {
		s.logger.WithContext(ctx).WithFields(logrus.Fields{
			"shortURLCode": request.ShortUrlCode,
		}).Info("Successfully deleted short URL Code from database")
	}

	return &ugen.DeleteShortURLResponse{
		Success: true,
		Message: "Short URL Code deleted successfully",
	}, nil
}

func (s *URLGenService) GetAllURLs(ctx context.Context, _ *emptypb.Empty) (*ugen.GetAllURLsResponse, error) {
	urls, err := s.repo.GetAllURLs(ctx)
	if err != nil {
		s.logger.WithContext(ctx).WithError(err).Error("Failed to retrieve all URLs")
		return nil, err
	}

	var urlResponses []*ugen.URLRecord
	for _, url := range urls {
		urlResponses = append(urlResponses, &ugen.URLRecord{
			ShortUrlCode: url.ShortURLCode,
			OriginalUrl:  url.OriginalURL,
			CreatedAt:    timestamppb.New(url.CreatedAt),
			ExpireTime:   timestamppb.New(*url.ExpiresAt),
		})
	}

	return &ugen.GetAllURLsResponse{
		Urls: urlResponses,
	}, nil
}
