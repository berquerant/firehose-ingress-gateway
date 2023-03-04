package amqp_test

import (
	"context"
	"testing"
	"time"

	"github.com/berquerant/firehose-ingress-gateway/pkg/amqp"
	"github.com/berquerant/firehose-ingress-gateway/server"
	queue "github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestGetMessage(t *testing.T) {
	const (
		connURI          = "amqp://user:pass@localhost:5672"
		exchangeName     = "default_exchange"
		exchangeKind     = "topic"
		queueName        = "default_queue"
		consumerName     = "default_consumer"
		consumerPrefetch = 1
		publishKey       = "default_key"
	)

	dial := func() *queue.Connection {
		conn, err := queue.Dial(connURI)
		if err != nil {
			t.Fatalf("failed to dial: %v", err)
		}
		return conn
	}

	deleteQueue := func() error {
		conn := dial()
		defer conn.Close()

		channel, err := conn.Channel()
		if err != nil {
			return err
		}

		_, err = channel.QueueDelete(queueName, false, false, false)
		return err
	}
	deleteQueue()
	defer deleteQueue()

	consumerAmqpConn := dial()
	defer consumerAmqpConn.Close()

	consumerConn := amqp.NewConn(
		consumerAmqpConn,
		amqp.WithExchangeName(exchangeName),
		amqp.WithExchangeKind(exchangeKind),
	)
	if err := consumerConn.Init(); err != nil {
		t.Fatalf("failed to init conn: %v", err)
	}
	defer consumerConn.Close()

	consumer := amqp.NewConsumer(
		consumerConn,
		amqp.WithQueueName(queueName),
		amqp.WithConsumerName(consumerName),
		amqp.WithPrefetchCount(consumerPrefetch),
	)
	if err := consumer.Init(); err != nil {
		t.Fatalf("failed to init consumer: %v", err)
	}
	defer consumer.Close()

	publisherAmqpConn := dial()
	defer publisherAmqpConn.Close()

	publisherConn, err := publisherAmqpConn.Channel()
	if err != nil {
		t.Fatalf("failed to open publisher channel: %v", err)
	}

	if err := publisherConn.ExchangeDeclare(
		exchangeName,
		exchangeKind,
		false, // durable
		false, // auto delete
		false, // internal
		false, // no wait
		nil,   // args
	); err != nil {
		t.Fatalf("failed to declare exchange: %v", err)
	}

	publish := func(body []byte) error {
		t.Logf("Publish %q", body)
		msg := queue.Publishing{
			DeliveryMode: queue.Persistent,
			Timestamp:    time.Now(),
			ContentType:  "text/plain",
			Body:         body,
		}
		return publisherConn.Publish(
			exchangeName,
			publishKey,
			false, // mandatory
			false, // immediate
			msg,
		)
	}

	// testcases
	var (
		first  = []byte("first")
		second = []byte("second")
	)

	for _, msg := range [][]byte{first, second} {
		if err := publish(msg); err != nil {
			t.Fatalf("failed to publish: %v", err)
		}
	}

	srv := amqp.New(consumer.Channel())

	t.Run("get first", func(t *testing.T) {
		msg, err := srv.GetMessage(context.TODO())
		assert.Nil(t, err)
		if !assert.NotNil(t, msg) {
			return
		}
		assert.Equal(t, first, msg.GetBody())
	})

	t.Run("get second", func(t *testing.T) {
		msg, err := srv.GetMessage(context.TODO())
		assert.Nil(t, err)
		if !assert.NotNil(t, msg) {
			return
		}
		assert.Equal(t, second, msg.GetBody())
	})

	t.Run("no messages", func(t *testing.T) {
		_, err := srv.GetMessage(context.TODO())
		assert.Equal(t, err, server.ErrMessageNotFound)
	})
}
