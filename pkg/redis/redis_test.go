package redis_test

import (
	"context"
	"testing"

	"github.com/berquerant/firehose-ingress-gateway/pkg/redis"
	"github.com/berquerant/firehose-ingress-gateway/server"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestGetMessage(t *testing.T) {
	client := goredis.NewClient(&goredis.Options{})

	_, err := client.Ping(context.TODO()).Result()
	if err != nil {
		t.Fatal("needs a running redis instance")
	}

	const key = "key"
	cleanRedis := func() error {
		_, err := client.Del(context.TODO(), key).Result()
		return err
	}

	assert.Nil(t, cleanRedis())
	defer cleanRedis()

	var (
		first  = []byte("first")
		second = []byte("second")
	)
	_, err = client.LPush(context.TODO(), key, first, second).Result()
	assert.Nil(t, err)

	srv := redis.New(client, key)

	t.Run("get first", func(t *testing.T) {
		msg, err := srv.GetMessage(context.TODO())
		assert.Nil(t, err)
		assert.Equal(t, first, msg.GetBody())
	})

	t.Run("get second", func(t *testing.T) {
		msg, err := srv.GetMessage(context.TODO())
		assert.Nil(t, err)
		assert.Equal(t, second, msg.GetBody())
	})

	t.Run("no messages", func(t *testing.T) {
		_, err := srv.GetMessage(context.TODO())
		assert.Equal(t, err, server.ErrMessageNotFound)
	})
}
