package main

import (
	"context"
	"fmt"
	"time"

	"github.com/berquerant/firehose-ingress-gateway/pkg/amqp"
	"github.com/berquerant/firehose-ingress-gateway/server"
	"github.com/berquerant/firehose-proto/envx"
	queue "github.com/streadway/amqp"
)

type config struct {
	// Port is a port number that the server listens on.
	Port int `env:"PORT" envDefault:"20001"`
	// URI is AMQP URI.
	URI           string        `env:"AMQP_URI"`
	ExchangeName  string        `env:"EXCHANGE_NAME"`
	ExchangeKind  string        `env:"EXCHANGE_KIND"`
	Queue         string        `env:"QUEUE" envDefault:"fig-queue"`
	Consumer      string        `env:"CONSUMER" envDefault:"fig-consumer"`
	PrefetchCount int           `env:"PREFETCH_COUNT" envDefault:"1"`
	GetTimeout    time.Duration `env:"GET_TIMEOUT" envDefault:"1s"`
}

func main() {
	var cfg config

	if err := envx.NewRunner(cfg).Run(run); err != nil {
		panic(err)
	}
}

func run(ctx context.Context, cfg config) error {
	amqpConn, err := queue.Dial(cfg.URI)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer amqpConn.Close()

	conn := amqp.NewConn(
		amqpConn,
		amqp.WithExchangeName(cfg.ExchangeName),
		amqp.WithExchangeKind(cfg.ExchangeKind),
	)
	if err := conn.Init(); err != nil {
		return fmt.Errorf("failed to init conn: %w", err)
	}
	defer conn.Close()

	consumer := amqp.NewConsumer(
		conn,
		amqp.WithQueueName(cfg.Queue),
		amqp.WithConsumerName(cfg.Consumer),
		amqp.WithPrefetchCount(cfg.PrefetchCount),
	)
	if err := consumer.Init(); err != nil {
		return fmt.Errorf("failed to init consumer: %w", err)
	}
	defer consumer.Close()

	messenger := amqp.New(
		consumer.Channel(),
		amqp.WithGetTimeout(cfg.GetTimeout),
	)
	srv := server.New(messenger)
	if err := server.Run(ctx, srv, cfg.Port); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	return nil
}
