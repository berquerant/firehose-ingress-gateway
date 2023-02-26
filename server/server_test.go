package server_test

import (
	"context"
	"errors"
	"testing"

	"github.com/berquerant/firehose-ingress-gateway/server"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockMessenger struct {
	msg server.Message
	err error
}

func (m *mockMessenger) GetMessage(context.Context) (server.Message, error) {
	return m.msg, m.err
}

func TestGetMessage(t *testing.T) {
	for _, tc := range []struct {
		title string
		err   error
		code  codes.Code
	}{
		{
			title: "canceled",
			err:   context.Canceled,
			code:  codes.Canceled,
		},
		{
			title: "deadline exceeded",
			err:   context.DeadlineExceeded,
			code:  codes.DeadlineExceeded,
		},
		{
			title: "no messages",
			err:   server.ErrMessageNotFound,
			code:  codes.NotFound,
		},
		{
			title: "internal error",
			err:   errors.New("InternalError"),
			code:  codes.Internal,
		},
	} {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			messenger := &mockMessenger{
				err: tc.err,
			}
			srv := server.New(messenger)
			_, err := srv.GetMessage(context.TODO(), nil)
			t.Logf("got err: %v", err)
			got := status.Code(err)
			assert.Equal(t, tc.code, got)
		})
	}

	t.Run("got message", func(t *testing.T) {
		var (
			id      = []byte("id")
			message = []byte("message")
		)
		messenger := &mockMessenger{
			msg: server.NewMessage(id, message),
		}
		srv := server.New(messenger)
		got, err := srv.GetMessage(context.TODO(), nil)
		assert.Nil(t, err)
		assert.Equal(t, id, got.GetId())
		assert.Equal(t, message, got.GetBody())
	})
}
