package testhelpers

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	redis "github.com/redis/go-redis/v9"
	testcontainers "github.com/testcontainers/testcontainers-go"
	postgrestest "github.com/testcontainers/testcontainers-go/modules/postgres"
	redistest "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	*postgrestest.PostgresContainer
	ConnectionString string
}

func CreatePostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	pgContainer, err := postgrestest.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgrestest.WithInitScripts(filepath.Join("testdata", "init-db.sql")),
		postgrestest.WithDatabase("urls"),
		postgrestest.WithUsername("test_user"),
		postgrestest.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}

	return &PostgresContainer{
		PostgresContainer: pgContainer,
		ConnectionString:  connStr,
	}, nil
}

type RedisContainer struct {
	*redistest.RedisContainer
	RedisOpts *redis.Options
}

func CreateRedisContainer(ctx context.Context) (*RedisContainer, error) {
	redisContainer, err := redistest.RunContainer(ctx,
		testcontainers.WithImage("redis:7-alpine"),
		redistest.WithSnapshotting(10, 1),
		redistest.WithLogLevel(redistest.LogLevelVerbose),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}

	host, err := redisContainer.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		return nil, err
	}

	return &RedisContainer{
		RedisContainer: redisContainer,
		RedisOpts: &redis.Options{
			Addr:     fmt.Sprintf("%s:%s", host, port.Port()),
			Password: "",
			DB:       0,
		},
	}, nil
}
