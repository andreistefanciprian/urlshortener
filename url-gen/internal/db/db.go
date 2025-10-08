package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type URLStore interface {
	CreateShortURL(ctx context.Context, originalURL, shortURLCode string, expireTime time.Time) error
	Delete(ctx context.Context, shortURLCode string) error
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

const insertShortURLQuery = `
		INSERT INTO short_links (code, original_url, expires_at) 
		VALUES ($1, $2, $3)`

func (r *PostgresURLStore) CreateShortURL(ctx context.Context, originalURL, shortURLCode string, expireTime time.Time) error {

	// SQL query to insert new short link
	_, err := r.db.Exec(ctx, insertShortURLQuery, shortURLCode, originalURL, expireTime)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"shortURLCode": shortURLCode,
			"originalURL":  originalURL,
		}).Error("Failed to create short URL in database")
		return fmt.Errorf("failed to create short URL: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"shortURLCode": shortURLCode,
		"originalURL":  originalURL,
		"expiresAt":    expireTime,
	}).Infof("Successfully created short URL")

	return nil
}

const deleteShortURLQuery = `
		DELETE FROM short_links 
		WHERE code = $1`

func (r *PostgresURLStore) Delete(ctx context.Context, shortURLCode string) error {
	// SQL query to delete a short link by code
	_, err := r.db.Exec(ctx, deleteShortURLQuery, shortURLCode)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"shortURLCode": shortURLCode,
		}).Error("Failed to delete short URL from database")
		return fmt.Errorf("failed to delete short URL: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"shortURLCode": shortURLCode,
	}).Info("Successfully deleted short URL")

	return nil
}
