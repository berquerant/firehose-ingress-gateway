package redis

import (
	"context"
	"errors"

	"github.com/berquerant/firehose-ingress-gateway/server"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

func New(client *goredis.Client, key string) *Server {
	return &Server{
		client: client,
		key:    key,
	}
}

type Server struct {
	client *goredis.Client
	key    string
}

func (s *Server) GetMessage(ctx context.Context) (server.Message, error) {
	msg, err := s.client.RPop(ctx, s.key).Bytes()

	switch {
	case errors.Is(err, goredis.Nil):
		return nil, server.ErrMessageNotFound
	case err != nil:
		return nil, err
	default:
		var id [16]byte = uuid.New()
		return server.NewMessage(id[:], msg), nil
	}
}
