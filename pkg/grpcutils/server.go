package grpcutils

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daneshvar/go-logger"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

type Server struct {
	log              *logger.Logger
	grpcServer       *grpc.Server
	grpcHealthServer *grpc.Server
	healthServer     *health.Server
	HealthAddr       string
	healthIsEnabled  bool
	RegisterService  func(context.Context, *Server) (closeFn func(), err error)

	HttpServerRun          func(ctx context.Context) error
	HttpServerGracefulStop func()

	healthServiceName string
}

func NewServer(log *logger.Logger, serverName string, disableHealth bool) *Server {
	return &Server{
		log:               log,
		healthIsEnabled:   !disableHealth,
		HealthAddr:        "localhost:9000",
		healthServiceName: fmt.Sprintf("grpc.health.v1.%s", serverName),
	}
}

func (s *Server) ServerGRPC() *grpc.Server {
	return s.grpcServer
}

func (s *Server) healthServerInit() {
	if s.healthIsEnabled {
		s.healthServer = health.NewServer()
		s.SetServingStatus(grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	}
}

func (s *Server) healthServerRun() error {
	if !s.healthIsEnabled {
		return nil
	}

	s.grpcHealthServer = grpc.NewServer()

	grpc_health_v1.RegisterHealthServer(s.grpcHealthServer, s.healthServer)

	hln, err := net.Listen("tcp", s.HealthAddr)
	if err != nil {
		s.log.Errorv("gRPC Health server: failed to listen", "error", err)
		os.Exit(2)
	}
	s.log.Infof("gRPC health server serving at %s", s.HealthAddr)

	return s.grpcHealthServer.Serve(hln)
}

func (s *Server) grpcServerRun(ctx context.Context, addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		s.log.Errorv("gRPC server: failed to listen", "error", err)
		os.Exit(2)
	}

	s.grpcServer = grpc.NewServer(
		// MaxConnectionAge is just to avoid long connection, to facilitate load balancing
		// MaxConnectionAgeGrace will torn them, default to infinity
		grpc.KeepaliveParams(keepalive.ServerParameters{MaxConnectionAge: 2 * time.Minute}),
		// grpc.StatsHandler(&ocgrpc_propag.ServerHandler{}),
		grpc.UnaryInterceptor(withUnaryServerLogger(s.log)),
		grpc.StreamInterceptor(withStreamServerLogger(s.log)),
	)

	if closeFn, err := s.RegisterService(ctx, s); err != nil {
		s.log.Errorv("gRPC server: failed to register", "error", err)
		os.Exit(2)
	} else if closeFn != nil {
		defer closeFn()
	}

	s.log.Infof("gRPC server serving at %s", addr)

	s.SetServingStatus(grpc_health_v1.HealthCheckResponse_SERVING)

	return s.grpcServer.Serve(ln)
}

func (s *Server) SetServingStatus(servingStatus grpc_health_v1.HealthCheckResponse_ServingStatus) {
	if s.healthIsEnabled {
		s.healthServer.SetServingStatus(s.healthServiceName, servingStatus)
	}
}

func (s *Server) Run(ctx context.Context, addr string) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM) // syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT
	defer signal.Stop(interruptChan)

	s.healthServerInit()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(s.healthServerRun)
	g.Go(func() error { return s.grpcServerRun(ctx, addr) })

	if s.HttpServerRun != nil {
		g.Go(func() error { return s.HttpServerRun(ctx) })
	}

	select {
	case <-interruptChan:
		break
	case <-ctx.Done():
		break
	}

	s.log.Warn("received shutdown signal")

	cancel()

	s.SetServingStatus(grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	if s.grpcHealthServer != nil {
		s.grpcHealthServer.GracefulStop()
	}

	if s.HttpServerGracefulStop != nil {
		s.HttpServerGracefulStop()
	}

	if err := g.Wait(); err != nil {
		s.log.Errorv("server returning an error", "error", err)
		os.Exit(2)
	}
}
