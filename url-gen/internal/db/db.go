package db

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type URLStore interface {
	CreateShortURL(ctx context.Context, originalURL, shortURLCode string, expireTime time.Time) error
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

	// Set connection pool settings
	poolConfig.MaxConns = 30
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30

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
		User:     getEnvOrDefault("DB_USER", "url_gen_user"),
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
