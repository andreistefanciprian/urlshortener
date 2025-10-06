package testhelpers

import (
	"context"
	"path/filepath"
	"time"

	testcontainers "github.com/testcontainers/testcontainers-go"
	postgrestest "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	*postgrestest.PostgresContainer
	ConnectionString string
}

func CreatePostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	pgContainer, err := postgrestest.Run(ctx,
		"postgres:15-alpine",
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
