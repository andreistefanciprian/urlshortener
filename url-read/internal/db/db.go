package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type URLStore interface {
	GetLongURL(ctx context.Context, shortURLCode string) (*URLRecord, error)
	Close() error
}

type PostgresURLStore struct {
	logger *logrus.Logger
	db     *pgxpool.Pool
}

// NewPostgresURLStore creates a new repository instance with database connection
func NewPostgresURLStore(logger *logrus.Logger, connectionString string) (*PostgresURLStore, error) {
	// Initialize database connection
	db, err := initDB(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &PostgresURLStore{
		logger: logger,
		db:     db,
	}, nil
}

func (r *PostgresURLStore) Close() error {
	r.db.Close()
	return nil
}

type URLRecord struct {
	OriginalURL string
	ExpiresAt   *time.Time
}

const getLongURLQuery = `
		SELECT original_url, expires_at
		FROM short_links 
		WHERE code = $1 AND (expires_at IS NULL OR expires_at > NOW())`

var ErrShortURLCodeNotFound = errors.New("short URL code not found")

func (r *PostgresURLStore) GetLongURL(ctx context.Context, shortURLCode string) (*URLRecord, error) {
	// SQL query to fetch the original URL by short code
	var response URLRecord
	err := r.db.QueryRow(ctx, getLongURLQuery, shortURLCode).Scan(&response.OriginalURL, &response.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.WithFields(logrus.Fields{
				"shortURLCode": shortURLCode,
			}).Info("Short URL not found in database")
			return nil, ErrShortURLCodeNotFound
		}
		r.logger.WithError(err).WithFields(logrus.Fields{
			"shortURLCode": shortURLCode,
		}).Error("Failed to fetch original URL from database")
		return nil, fmt.Errorf("failed to retrieve URL: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"shortURLCode": shortURLCode,
		"originalURL":  response.OriginalURL,
		"expiresAt":    response.ExpiresAt,
	}).Infof("Successfully fetched original URL from database")

	return &URLRecord{
		OriginalURL: response.OriginalURL,
		ExpiresAt:   response.ExpiresAt,
	}, nil
}
