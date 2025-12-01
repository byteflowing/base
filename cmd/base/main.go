package main

import (
	"flag"

	"github.com/byteflowing/base/pkg/utils/slicex"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	"google.golang.org/grpc"

	"github.com/byteflowing/base/app/geo"
	"github.com/byteflowing/base/app/global_id"
	"github.com/byteflowing/base/app/maps"
	"github.com/byteflowing/base/app/message"
	"github.com/byteflowing/base/app/user"
	"github.com/byteflowing/base/pkg/logx"
	"github.com/byteflowing/base/singleton"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	geov1 "github.com/byteflowing/proto/gen/go/geo/v1"
	globalidv1 "github.com/byteflowing/proto/gen/go/global_id/v1"
	mapsv1 "github.com/byteflowing/proto/gen/go/maps/v1"
	msgv1 "github.com/byteflowing/proto/gen/go/msg/v1"
	userv1 "github.com/byteflowing/proto/gen/go/user/v1"
)

func main() {
	configPath := flag.String("config", "./base.dev.yaml", "base -config base.yaml")
	flag.Parse()
	cfg := singleton.NewConfig(*configPath)
	logx.Init(cfg.Log)

	server := NewGrpcServer(cfg, getServices(cfg.Services))
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

func RegisterMapService(c *configv1.Config, grpcServer *grpc.Server) {
	srv := maps.NewOnce(c)
	mapsv1.RegisterMapServiceServer(grpcServer, srv)
}

func RegisterUserService(c *configv1.Config, grpcServer *grpc.Server) {
	srv := user.NewOnce(c)
	userv1.RegisterUserServiceServer(grpcServer, srv)
}

func getServices(ss []enumsv1.SupportedService) []RegisterFn {
	var fn []RegisterFn
	ss = slicex.Unique(ss)
	for _, s := range ss {
		switch s {
		case enumsv1.SupportedService_SUPPORTED_SERVICE_GEO:
			fn = append(fn, RegisterGeoService)
		case enumsv1.SupportedService_SUPPORTED_SERVICE_GLOBAL_ID:
			fn = append(fn, RegisterGlobalIDService)
		case enumsv1.SupportedService_SUPPORTED_SERVICE_MAPS:
			fn = append(fn, RegisterMapService)
		case enumsv1.SupportedService_SUPPORTED_SERVICE_MESSAGE:
			fn = append(fn, RegisterMessageService)
		case enumsv1.SupportedService_SUPPORTED_SERVICE_USER:
			fn = append(fn, RegisterUserService)
		}
	}
	return fn
}
