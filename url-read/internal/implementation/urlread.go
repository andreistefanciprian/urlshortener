package urlread

import (
	"context"
	"time"

	cache "github.com/andreistefanciprian/urlshortener/url-read/internal/cache"
	repo "github.com/andreistefanciprian/urlshortener/url-read/internal/db"
	uread "github.com/andreistefanciprian/urlshortener/url-read/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type URLReadService struct {
	logger *logrus.Logger
	repo   repo.URLStore
	uread.UnimplementedURLReaderServer
	cache cache.URLCache
}

func NewURLReadService(logger *logrus.Logger, repository repo.URLStore, cache cache.URLCache) *URLReadService {
	return &URLReadService{
		logger: logger,
		repo:   repository,
		cache:  cache,
	}
}

func (s *URLReadService) GetLongURL(ctx context.Context, request *uread.LongURLRequest) (*uread.LongURLResponse, error) {

	// Retrieve the original URL from Cache
	cacheResponse, err := s.cache.Get(ctx, request.ShortUrl)
	if err != nil {
		s.logger.WithContext(ctx).WithError(err).WithFields(logrus.Fields{
			"shortURLCode": request.ShortUrl,
		}).Errorf("Failed to retrieve original URL from cache: %v", err)
	} else if cacheResponse != nil {
		s.logger.WithContext(ctx).WithFields(logrus.Fields{
			"shortURLCode": request.ShortUrl,
			"originalURL":  cacheResponse.LongURL,
			"expiration":   cacheResponse.ExpiresAt,
		}).Debug("Successfully retrieved original URL from cache")
		var expiration *timestamppb.Timestamp
		if !cacheResponse.ExpiresAt.IsZero() {
			expiration = timestamppb.New(cacheResponse.ExpiresAt)
		}
		return &uread.LongURLResponse{
			LongUrl:    cacheResponse.LongURL,
			Expiration: expiration,
		}, nil
	}

	// If Cache MISS
	// Retrieve the original URL from the database (context will be used by db layer)
	response, err := s.repo.GetLongURL(ctx, request.ShortUrl)
	if err != nil {
		s.logger.WithContext(ctx).WithError(err).WithFields(logrus.Fields{
			"shortURLCode": request.ShortUrl,
		}).Errorf("Failed to retrieve original URL from database: %v", err)
		return nil, err
	}

	s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"shortURLCode": request.ShortUrl,
		"originalURL":  response.OriginalURL,
		"expiration":   response.ExpiresAt,
	}).Debug("Successfully retrieved original URL from database")
	var expiration *timestamppb.Timestamp
	if response.ExpiresAt != nil {
		expiration = timestamppb.New(*response.ExpiresAt)
	}

	// Only cache the URL if it hasn't expired
	if response.ExpiresAt == nil || !response.ExpiresAt.Before(time.Now()) {
		// Store the original URL in Cache for future requests
		var expiresAt time.Time
		if response.ExpiresAt != nil {
			expiresAt = *response.ExpiresAt
		}
		err = s.cache.Set(ctx, request.ShortUrl, cache.URLCacheEntry{
			LongURL:   response.OriginalURL,
			ExpiresAt: expiresAt,
		})
		if err != nil {
			s.logger.WithContext(ctx).WithError(err).WithFields(logrus.Fields{
				"shortURLCode": request.ShortUrl,
				"originalURL":  response.OriginalURL,
				"expiration":   response.ExpiresAt,
			}).Errorf("Failed to store original URL in cache: %v", err)
			// Proceeding without failing the request, as we have the data from DB
		}
	} else {
		s.logger.WithContext(ctx).WithFields(logrus.Fields{
			"shortURLCode": request.ShortUrl,
			"originalURL":  response.OriginalURL,
			"expiration":   response.ExpiresAt,
		}).Warn("Skipping cache storage for expired URL")
	}

	return &uread.LongURLResponse{
		LongUrl:    response.OriginalURL,
		Expiration: expiration,
	}, nil
}
