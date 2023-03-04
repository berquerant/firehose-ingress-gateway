package amqp

import (
	"context"
	"time"

	"github.com/berquerant/firehose-ingress-gateway/server"
	queue "github.com/streadway/amqp"
)

//go:generate go run github.com/berquerant/goconfig@v0.2.0 -field "GetTimeout time.Duration" -prefix Server -option -output amqp_config_generated.go

func New(ch <-chan queue.Delivery, opt ...ServerConfigOption) *Server {
	config := NewServerConfigBuilder().
		GetTimeout(time.Second).
		Build()
	config.Apply(opt...)
	return &Server{
		ch:     ch,
		config: config,
	}
}

type Server struct {
	ch     <-chan queue.Delivery
	config *ServerConfig
}

func (s *Server) GetMessage(ctx context.Context) (server.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, s.config.GetTimeout.Get())
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, server.ErrMessageNotFound
	case msg := <-s.ch:
		if err := msg.Ack(false); err != nil {
			return nil, err
		}
		return server.NewMessage(
			[]byte(msg.MessageId),
			msg.Body,
		), nil
	}
}
