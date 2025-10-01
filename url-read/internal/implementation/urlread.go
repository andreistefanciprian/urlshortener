package urlread

import (
	"context"

	cache "github.com/andreistefanciprian/urlshortener/url-read/internal/cache"
	repo "github.com/andreistefanciprian/urlshortener/url-read/internal/db"
	uread "github.com/andreistefanciprian/urlshortener/url-read/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Implementation struct {
	logger *logrus.Logger
	repo   repo.URLRepository
	uread.UnimplementedURLReaderServer
	cache cache.URLCacher
}

func NewUrlReadImplementation(logger *logrus.Logger, repository repo.URLRepository, cache cache.URLCacher) *Implementation {
	return &Implementation{
		logger: logger,
		repo:   repository,
		cache:  cache,
	}
}

func (s *Implementation) GetLongURL(ctx context.Context, request *uread.LongURLRequest) (*uread.LongURLResponse, error) {

	// Retrieve the original URL from Cache
	cacheResponse, err := s.cache.Get(ctx, request.ShortUrl)
	if err != nil {
		s.logger.WithContext(ctx).WithError(err).WithFields(logrus.Fields{
			"shortCode": request.ShortUrl,
		}).Errorf("Failed to retrieve original URL from cache: %v", err)
	} else if cacheResponse != nil {
		s.logger.WithContext(ctx).WithFields(logrus.Fields{
			"shortCode":   request.ShortUrl,
			"originalURL": cacheResponse.LongURL,
			"expiration":  cacheResponse.ExpiresAt,
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
			"shortCode": request.ShortUrl,
		}).Errorf("Failed to retrieve original URL from database: %v", err)
		return nil, err
	}

	s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"shortCode":   request.ShortUrl,
		"originalURL": response.OriginalURL,
		"expiration":  response.ExpiresAt,
	}).Debug("Successfully retrieved original URL from database")
	var expiration *timestamppb.Timestamp
	if response.ExpiresAt != nil {
		expiration = timestamppb.New(*response.ExpiresAt)
	}

	// Store the original URL in Cache for future requests
	err = s.cache.Set(ctx, request.ShortUrl, cache.CachedURL{
		LongURL:   response.OriginalURL,
		ExpiresAt: *response.ExpiresAt,
	})
	if err != nil {
		s.logger.WithContext(ctx).WithError(err).WithFields(logrus.Fields{
			"shortCode":   request.ShortUrl,
			"originalURL": response.OriginalURL,
			"expiration":  response.ExpiresAt,
		}).Errorf("Failed to store original URL in cache: %v", err)
		// Proceeding without failing the request, as we have the data from DB
	}

	return &uread.LongURLResponse{
		LongUrl:    response.OriginalURL,
		Expiration: expiration,
	}, nil
}

// func validateExpiresAt(expiration *time.Time) error {
// 	if expiration != nil && expiration.Before(time.Now()) {
// 		return fmt.Errorf("expiration time cannot be in the past")
// 	}
// 	return nil
// }
