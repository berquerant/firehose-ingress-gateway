package server

import (
	"context"
	"errors"

	"github.com/berquerant/firehose-proto/ingress"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Message interface {
	GetID() []byte
	GetBody() []byte
}

func NewMessage(id, body []byte) *MessageImpl {
	return &MessageImpl{
		ID:   id,
		Body: body,
	}
}

type MessageImpl struct {
	ID   []byte
	Body []byte
}

func (m *MessageImpl) GetID() []byte   { return m.ID }
func (m *MessageImpl) GetBody() []byte { return m.Body }

var (
	ErrMessageNotFound = errors.New("MessageNotFound")
)

type Messenger interface {
	GetMessage(ctx context.Context) (Message, error)
}

type Server struct {
	messenger Messenger
	ingress.UnimplementedIngressGatewayServiceServer
}

func (g *Server) GetMessage(ctx context.Context, _ *ingress.GetMessageRequest) (*ingress.GetMessageResponse, error) {
	msg, err := g.messenger.GetMessage(ctx)

	switch {
	case err == nil:
		return &ingress.GetMessageResponse{
			Body: msg.GetBody(),
			Id:   msg.GetID(),
		}, nil
	case errors.Is(err, context.DeadlineExceeded):
		return nil, status.Errorf(codes.DeadlineExceeded, "Context deadline exceeded: %v", err)
	case errors.Is(err, context.Canceled):
		return nil, status.Errorf(codes.Canceled, "Context canceled: %v", err)
	case errors.Is(err, ErrMessageNotFound):
		return nil, status.Errorf(codes.NotFound, "No messages: %v", err)
	default:
		return nil, status.Errorf(codes.Internal, "Internal error: %v", err)
	}
}

func New(messenger Messenger) *Server {
	return &Server{
		messenger: messenger,
	}
}
