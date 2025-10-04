package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
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

func (r *PostgresURLStore) GetLongURL(ctx context.Context, shortURLCode string) (*URLRecord, error) {
	// SQL query to fetch the original URL by short code
	var response URLRecord
	err := r.db.QueryRow(ctx, getLongURLQuery, shortURLCode).Scan(&response.OriginalURL, &response.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.WithFields(logrus.Fields{
				"shortURLCode": shortURLCode,
			}).Info("Short URL not found in database")
			return nil, fmt.Errorf("short URL '%s' not found", shortURLCode)
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

// DBConfig holds database configuration parameters
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DSNString constructs the Data Source Name string from the configuration
func (c *DBConfig) DSNString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// initDB collects database parameters from environment variables and creates connection pool
func initDB(connectionString string) (*pgxpool.Pool, error) {

	// Create connection pool configuration
	poolConfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Set connection pool settings optimized for read operations
	poolConfig.MaxConns = 50                      // Higher for read concurrency
	poolConfig.MinConns = 10                      // Keep more connections warm for reads
	poolConfig.MaxConnLifetime = time.Hour * 2    // Longer lifetime for read replicas
	poolConfig.MaxConnIdleTime = time.Minute * 15 // Shorter idle time to free up connections faster

	// Create connection pool
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// GetDBConfigFromEnv collects database configuration from environment variables
func GetDBConfigFromEnv() DBConfig {
	config := DBConfig{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvAsIntOrDefault("DB_PORT", 5432),
		User:     getEnvOrDefault("DB_USER", "url_read_user"),
		Password: getEnvOrDefault("DB_PASSWORD", "Auth123"),
		DBName:   getEnvOrDefault("DB_NAME", "urls"),
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}
	return config
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsIntOrDefault returns environment variable as int or default if not set or invalid
func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
