package main

import (
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/byteflowing/base/pkg/logx"
	"github.com/byteflowing/base/pkg/signalx"
	"github.com/byteflowing/base/singleton"
	"github.com/byteflowing/base/version"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
)

// RegisterFn 绑定grpc实现到grpc.Server上
type RegisterFn func(cfg *configv1.Config, server *grpc.Server)

type Server struct {
	cfg       *configv1.Config
	server    *grpc.Server
	registers []RegisterFn
	signal    *signalx.SignalListener
}

func NewGrpcServer(
	cfg *configv1.Config,
	registers []RegisterFn,
	opts ...grpc.ServerOption,
) *Server {
	server := grpc.NewServer(opts...)
	signal := signalx.NewSignalListener(cfg.Server.WaitForShutdown.AsDuration())
	return &Server{
		cfg:       cfg,
		server:    server,
		registers: registers,
		signal:    signal,
	}
}

func (s *Server) Start() {
	lis, err := net.Listen("tcp", s.cfg.Server.Addr)
	if err != nil {
		logx.Fatal("listen grpc server fail", zap.Error(err))
	}
	for _, register := range s.registers {
		register(s.cfg, s.server)
	}
	logx.Info("grpc server started", zap.String("addr", s.cfg.Server.Addr))
	version.PrintVersion()
	if err := s.server.Serve(lis); err != nil {
		logx.Fatal("grpc server failed to serve", zap.Error(err))
	}
	logx.Info("grpc server stopped", zap.String("addr", s.cfg.Server.Addr))
}

func (s *Server) Stop() {
	logx.Info("grpc graceful stop method is called, so stopping...", zap.String("addr", s.cfg.Server.Addr))
	s.server.GracefulStop()
}

func (s *Server) Spin() {
	s.signal.Add(singleton.GetStarterMgr())
	s.signal.Add(s)
	s.signal.Listen()
}
