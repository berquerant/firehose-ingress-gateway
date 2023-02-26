package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/berquerant/firehose-proto/grpcx"
	"github.com/berquerant/firehose-proto/ingress"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

func Run(ctx context.Context, srv *Server, port int) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	opts := []grpc.ServerOption{
		grpcmiddleware.WithUnaryServerChain(
			grpcx.NewBaseUnaryServerInterceptors()...,
		),
	}

	opt, err := grpcx.NewServerCredentialsFromEnv()
	if err != nil && !errors.Is(err, grpcx.ErrNoCredentials) {
		return fmt.Errorf("Failed to read credentials %w", err)
	}
	if opt != nil {
		opts = append(opts, opt)
	}

	s := grpc.NewServer(opts...)
	ingress.RegisterIngressGatewayServiceServer(s, srv)

	runner := grpcx.NewRunner(
		grpcx.NewServer(
			s,
			port,
		),
	)
	runner.Run(ctx)
	return runner.Err()
}
