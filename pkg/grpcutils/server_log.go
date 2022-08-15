package grpcutils

import (
	"context"

	"github.com/daneshvar/go-logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func withUnaryServerLogger(log *logger.Logger) grpc.UnaryServerInterceptor {
	log1 := log.GetLogger("grpc").Skip(1)
	log2 := log.GetLogger("grpc").Skip(2)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		panicked := true

		defer func() {
			if r := recover(); r != nil || panicked {
				log2.Errorv("Panic on Request", "method", info.FullMethod, "error", r)
				err = status.Error(codes.Internal, "Internal Error")
			}
		}()

		resp, err = handler(ctx, req)
		panicked = false

		if err != nil {
			log1.Warnv("Error on Request", "method", info.FullMethod, "error", err)
		}
		return resp, err
	}
}

func withStreamServerLogger(log *logger.Logger) grpc.StreamServerInterceptor {
	log1 := log.GetLogger("grpc").Skip(1)
	log2 := log.GetLogger("grpc").Skip(2)

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		panicked := true

		defer func() {
			if r := recover(); r != nil || panicked {
				log2.Errorv("Panic on Request", "method", info.FullMethod, "error", r)
				err = status.Errorf(codes.Internal, "%v", r)
			}
		}()

		err = handler(srv, ss)
		panicked = false

		if err != nil {
			log1.Warnv("Error on Request", "method", info.FullMethod, "error", err)
		}
		return err
	}
}
