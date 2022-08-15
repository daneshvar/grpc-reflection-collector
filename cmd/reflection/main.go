package main

import (
	"context"

	"github.com/daneshvar/go-logger"
	loggerinflux "github.com/daneshvar/go-logger-influx"
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
	loggerConfig(log, logger, &cfg.Logger)

	s := grpcutils.NewServer(log, "reflection", cfg.DisableHealth)
	s.RegisterService = func(ctx context.Context, s *grpcutils.Server) (func(), error) {
		refServer, err := reflection.NewServerReflectionServer(ctx, log.GetLogger("reflect"), cfg.Services, cfg.Ignores, cfg.Cache)
		if err != nil {
			log.Errorf("failed to create server: %v", err)
			return nil, err
		}
		rpb.RegisterServerReflectionServer(s.ServerGRPC(), refServer)
		return nil, nil
	}

	s.Run(context.Background(), cfg.Addr)
}

func loggerConfig(log *logger.Logger, core *logger.Core, cfg *LoggerConfig) {
	writers := make([]*logger.Writer, 0)

	if cfg.Console != nil {
		consoleWr, err := logger.ConsoleWriterWithConfig(cfg.Console)
		if err != nil {
			log.Fatalv("logger console config", "error", err)
		}
		writers = append(writers, consoleWr)
	}

	if cfg.Influx != nil {
		influxWr, err := loggerinflux.WriterWithConfig(cfg.Influx)
		if err != nil {
			log.Fatalv("logger influx config", "error", err)
		}
		writers = append(writers, influxWr)
	}

	if len(writers) > 0 {
		core.Config(writers...)
	}
}
