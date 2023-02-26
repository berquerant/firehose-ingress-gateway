package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/berquerant/firehose-ingress-gateway/pkg/redis"
	"github.com/berquerant/firehose-ingress-gateway/server"
	"github.com/caarlos0/env/v7"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	var (
		cfg  config
		opts = env.Options{
			OnSet: func(tag string, value any, isDefault bool) {
				logger.Info("Read env",
					zap.String("name", tag),
					zap.Any("value", value),
					zap.Bool("default", isDefault),
				)
			},
		}
	)

	if err := env.Parse(&cfg, opts); err != nil {
		logger.Fatal("Failed to read env",
			zap.Error(err),
		)
	}
	if err := run(&cfg); err != nil {
		logger.Fatal("Run",
			zap.Error(err),
		)
	}
}

func run(cfg *config) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	rdb := goredis.NewClient(&goredis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("Failed to connect to redis: %w", err)
	}

	messenger := redis.New(rdb, cfg.RedisQueueKey)
	srv := server.New(messenger)
	if err := server.Run(ctx, srv, cfg.Port); err != nil {
		return fmt.Errorf("Run: %w", err)
	}
	return nil
}
