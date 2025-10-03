package testhelpers

import (
	"context"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
	testcontainers "github.com/testcontainers/testcontainers-go"
	redistest "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

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
