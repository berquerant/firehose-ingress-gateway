package main_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/berquerant/firehose-proto/ingress"
	"github.com/berquerant/firehose-test/grpctest"
	"github.com/berquerant/firehose-test/tempdir"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRun(t *testing.T) {
	const (
		redisQueueKey = "figr-test"
	)
	var (
		rdb        = goredis.NewClient(&goredis.Options{})
		cleanRedis = func() error {
			_, err := rdb.Del(context.TODO(), redisQueueKey).Result()
			return err
		}
		push = func(value []byte) error {
			_, err := rdb.LPush(context.TODO(), redisQueueKey, value).Result()
			return err
		}
		firstValue  = []byte("first")
		secondValue = []byte("second")
	)

	if err := func() error {
		if _, err := rdb.Ping(context.TODO()).Result(); err != nil {
			return err
		}
		if err := cleanRedis(); err != nil {
			return err
		}
		for _, v := range [][]byte{
			firstValue,
			secondValue,
		} {
			if err := push(v); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		t.Fatalf("Failed to prepare redis: %v", err)
	}
	defer cleanRedis()

	dir := tempdir.New("firehose-ingress-gateway-redis")
	defer dir.Close()

	os.Setenv("REDIS_QUEUE_KEY", redisQueueKey)
	runner := grpctest.NewRunner(dir, grpctest.WithHealthWait(3*time.Second))
	defer runner.Close()
	assert.Nil(t, runner.Init(context.TODO()))

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
