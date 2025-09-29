package urlread

import (
	"context"

	repo "github.com/andreistefanciprian/urlshortener/url-read/internal/db"
	uread "github.com/andreistefanciprian/urlshortener/url-read/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Implementation struct {
	logger *logrus.Logger
	repo   repo.URLRepository
	uread.UnimplementedURLReaderServer
}

func NewUrlReadImplementation(logger *logrus.Logger, repository repo.URLRepository) *Implementation {
	return &Implementation{
		logger: logger,
		repo:   repository,
	}
}

func (s *Implementation) GetLongURL(ctx context.Context, request *uread.LongURLRequest) (*uread.LongURLResponse, error) {
	s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"shortCode": request.ShortUrl,
	}).Info("Processing long URL read request")

	// Check if context is cancelled
	if ctx.Err() != nil {
		s.logger.WithContext(ctx).Error("Request context cancelled")
		return nil, ctx.Err()
	}

	// Retrieve the original URL from the database (context will be used by db layer)
	response, err := s.repo.GetLongURL(ctx, request.ShortUrl)
	if err != nil {
		s.logger.WithContext(ctx).WithError(err).WithFields(logrus.Fields{
			"shortCode": request.ShortUrl,
		}).Error("Failed to retrieve original URL")
		return nil, err
	}

	s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"shortCode":   request.ShortUrl,
		"originalURL": response.OriginalURL,
	}).Info("Successfully retrieved original URL")
	var expiration *timestamppb.Timestamp
	if response.ExpiresAt != nil {
		expiration = timestamppb.New(*response.ExpiresAt)
	}
	return &uread.LongURLResponse{
		LongUrl:    response.OriginalURL,
		Expiration: expiration,
	}, nil
}
