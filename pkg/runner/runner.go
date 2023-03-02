package runner

import (
	"context"
	"os"
	"os/signal"

	"github.com/caarlos0/env/v7"
	"go.uber.org/zap"
)

type Runner[T any] struct {
	config T
}

func New[T any](config T) *Runner[T] {
	return &Runner[T]{
		config: config,
	}
}

func (r *Runner[T]) Run(ctx context.Context, f func(context.Context, T) error) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	opts := env.Options{
		OnSet: func(tag string, value any, isDefault bool) {
			logger.Info("Read env",
				zap.String("name", tag),
				zap.Any("value", value),
				zap.Bool("default", isDefault),
			)
		},
	}

	if err := env.Parse(&r.config, opts); err != nil {
		logger.Fatal("Failed to read env",
			zap.Error(err),
		)
		return err
	}

	return f(ctx, r.config)
}
