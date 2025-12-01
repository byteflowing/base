package main

import (
	"flag"

	"google.golang.org/grpc"

	"github.com/byteflowing/base/app/geo"
	"github.com/byteflowing/base/app/global_id"
	"github.com/byteflowing/base/app/message"
	"github.com/byteflowing/base/pkg/logx"
	"github.com/byteflowing/base/singleton"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	geov1 "github.com/byteflowing/proto/gen/go/geo/v1"
	globalidv1 "github.com/byteflowing/proto/gen/go/global_id/v1"
	msgv1 "github.com/byteflowing/proto/gen/go/msg/v1"
)

func main() {
	configPath := flag.String("config", "./grpc_base.dev.yaml", "grpc_base -config grpc_base.yaml")
	flag.Parse()
	cfg := singleton.NewConfig(*configPath)
	logx.Init(cfg.Log)
	server := NewGrpcServer(cfg, []RegisterFn{
		RegisterGeoService,
		RegisterMessageService,
		RegisterGlobalIDService,
	})
	server.Spin()
	_ = logx.Sync()
}

func RegisterGeoService(c *configv1.Config, grpcServer *grpc.Server) {
	srv := geo.NewOnce(c)
	geov1.RegisterGeoServiceServer(grpcServer, srv)
}

func RegisterGlobalIDService(c *configv1.Config, grpcServer *grpc.Server) {
	srv := global_id.NewOnce(c)
	globalidv1.RegisterGlobalIdServiceServer(grpcServer, srv)
}

func RegisterMessageService(c *configv1.Config, grpcServer *grpc.Server) {
	srv := message.NewOnce(c)
	msgv1.RegisterMessageServiceServer(grpcServer, srv)
}
