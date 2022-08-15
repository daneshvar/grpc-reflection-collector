package main

import (
	"context"

	"github.com/daneshvar/go-logger"
	"github.com/daneshvar/grpc-reflection-collector/internal/reflection"
	"github.com/daneshvar/grpc-reflection-collector/pkg/grpcutils"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

func main() {
	logger := logger.New()
	defer logger.Close()
	logger.RedirectStdLog("std")

	log := logger.GetLogger("boot")

	cfg := loadConfig(log)

	s := grpcutils.NewServer(log, "reflection", cfg.DisableHealth)
	s.RegisterService = func(ctx context.Context, s *grpcutils.Server) (func(), error) {
		refServer, err := reflection.NewServerReflectionServer(ctx, log.GetLogger("reflect"), cfg.Services, cfg.Ignores)
		if err != nil {
			log.Errorf("failed to create server: %v", err)
			return nil, err
		}
		rpb.RegisterServerReflectionServer(s.ServerGRPC(), refServer)
		return nil, nil
	}

	s.Run(context.Background(), cfg.Addr)
}
