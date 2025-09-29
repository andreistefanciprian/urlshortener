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

type ShortURLRepository interface {
	GetLongURL(ctx context.Context, shortCode string) (*LongURLRecord, error)
	Close() error
}

type MyShortURLRepository struct {
	logger *logrus.Logger
	db     *pgxpool.Pool
}

// NewMyShortURLRepository creates a new repository instance with database connection
func NewMyShortURLRepository(logger *logrus.Logger) (*MyShortURLRepository, error) {
	// Initialize database connection
	db, err := initDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &MyShortURLRepository{
		logger: logger,
		db:     db,
	}, nil
}

func (r *MyShortURLRepository) Close() error {
	r.db.Close()
	return nil
}

type LongURLRecord struct {
	OriginalURL string
	ExpiresAt   *time.Time
}

func (r *MyShortURLRepository) GetLongURL(ctx context.Context, shortCode string) (*LongURLRecord, error) {
	// SQL query to fetch the original URL by short code
	query := `
		SELECT original_url, expires_at
		FROM short_links 
		WHERE code = $1 AND (expires_at IS NULL OR expires_at > NOW())`

	var response LongURLRecord
	err := r.db.QueryRow(ctx, query, shortCode).Scan(&response.OriginalURL, &response.ExpiresAt)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"shortCode": shortCode,
		}).Error("Failed to fetch original URL from database")
		return nil, fmt.Errorf("failed to fetch original URL: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"shortCode":   shortCode,
		"originalURL": response.OriginalURL,
		"expiresAt":   response.ExpiresAt,
	}).Infof("Successfully fetched original URL")

	return &LongURLRecord{
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
func initDB() (*pgxpool.Pool, error) {

	// Get database configuration from environment variables
	config := getDBConfigFromEnv()

	// Create connection pool configuration
	poolConfig, err := pgxpool.ParseConfig(config.DSNString())
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

// getDBConfigFromEnv collects database configuration from environment variables
func getDBConfigFromEnv() DBConfig {
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
