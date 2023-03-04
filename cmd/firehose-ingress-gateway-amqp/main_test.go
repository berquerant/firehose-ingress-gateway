package main_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/berquerant/firehose-proto/ingress"
	"github.com/berquerant/firehose-test/grpctest"
	"github.com/berquerant/firehose-test/tempdir"
	queue "github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRun(t *testing.T) {
	const (
		uri          = "amqp://user:pass@localhost:5672"
		exchangeName = "default_main_exchange"
		exchangeKind = "topic"
		queueName    = "default_main_queue"
		publishKey   = "default_main_key"
		serverPort   = 20003
	)
	os.Setenv("AMQP_URI", uri)
	os.Setenv("EXCHANGE_NAME", exchangeName)
	os.Setenv("EXCHANGE_KIND", exchangeKind)
	os.Setenv("QUEUE", queueName)
	os.Setenv("GET_TIMEOUT", "1s")

	var (
		firstValue  = []byte("first")
		secondValue = []byte("second")
	)

	dial := func() *queue.Connection {
		conn, err := queue.Dial(uri)
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

	dir := tempdir.New("firehose-ingress-gateway-amqp")
	defer dir.Close()

	runner := grpctest.NewRunner(
		dir,
		grpctest.WithHealthWait(3*time.Second),
		grpctest.WithPort(serverPort),
	)
	defer runner.Close()
	assert.Nil(t, runner.Init(context.TODO()))

	{
		conn := dial()
		defer conn.Close()

		channel, err := conn.Channel()
		if err != nil {
			t.Fatalf("failed to open channel: %v", err)
		}

		publish := func(body []byte) error {
			t.Logf("Publish %q", body)
			msg := queue.Publishing{
				DeliveryMode: queue.Persistent,
				Timestamp:    time.Now(),
				ContentType:  "text/plain",
				Body:         body,
			}
			return channel.Publish(
				exchangeName,
				publishKey,
				false, // mandatory
				false, // immediate
				msg,
			)
		}

		for _, msg := range [][]byte{firstValue, secondValue} {
			if err := publish(msg); err != nil {
				t.Fatalf("failed to publish: %v", err)
			}
		}
	}

	client := ingress.NewIngressGatewayServiceClient(runner.Conn)

	t.Run("pop first value", func(t *testing.T) {
		resp, err := client.GetMessage(context.TODO(), new(ingress.GetMessageRequest))
		assert.Nil(t, err)
		assert.Equal(t, firstValue, resp.GetBody())
	})

	t.Run("pop second value", func(t *testing.T) {
		resp, err := client.GetMessage(context.TODO(), new(ingress.GetMessageRequest))
		assert.Nil(t, err)
		assert.Equal(t, secondValue, resp.GetBody())
	})

	t.Run("no messages", func(t *testing.T) {
		_, err := client.GetMessage(context.TODO(), new(ingress.GetMessageRequest))
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
}
