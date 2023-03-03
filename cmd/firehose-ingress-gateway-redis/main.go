package main

import (
	"context"
	"fmt"

	"github.com/berquerant/firehose-ingress-gateway/pkg/redis"
	"github.com/berquerant/firehose-ingress-gateway/server"
	"github.com/berquerant/firehose-proto/envx"
	goredis "github.com/redis/go-redis/v9"
)

type config struct {
	// Port is a port number that the server listens on.
	Port int `env:"PORT" envDefault:"20001"`
	// RedisAddr is host:port of redis server.
	RedisAddr string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	// RedisPass is password of redis user.
	RedisPass string `env:"REDIS_PASS" envDefault:""`
	// RedisDB is db of redis.
	RedisDB int `env:"REDIS_DB" envDefault:"0"`
	// RedisQueueKey is a key of the item that queues messages.
	RedisQueueKey string `env:"REDIS_QUEUE_KEY" envDefault:"figrq"`
}

func main() {
	var cfg config

	if err := envx.NewRunner(cfg).Run(run); err != nil {
		panic(err)
	}
}

func run(ctx context.Context, cfg config) error {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	messenger := redis.New(rdb, cfg.RedisQueueKey)
	srv := server.New(messenger)
	if err := server.Run(ctx, srv, cfg.Port); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	return nil
}
