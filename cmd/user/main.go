package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"

	"github.com/byteflowing/base/pkg/user"
	"github.com/byteflowing/base/pkg/utils"
	"github.com/byteflowing/base/version"
	"github.com/byteflowing/go-common/logx"
	"github.com/byteflowing/go-common/signalx"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	userv1 "github.com/byteflowing/proto/gen/go/services/user/v1"
	protovalidatemiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
)

func main() {
	configPath := flag.String("config", "./config.dev.yaml", "user -config config.yaml")
	flag.Parse()
	userImpl := NewWithConfig(*configPath)
	config := userImpl.GetConfig()
	logx.Init(config.LogConfig)
	version.PrintFullVersion()
	userService := newSrv(userImpl, config)
	signalListener := signalx.NewSignalListener(30 * time.Second)
	signalListener.Register(userService)
	signalListener.Listen()
	log.Println("user service exited")
	version.PrintFullVersion()
}

type srv struct {
	user      *user.Impl
	config    *configv1.Config
	grpServer *grpc.Server
}

func newSrv(user *user.Impl, config *configv1.Config) *srv {
	return &srv{
		user:   user,
		config: config,
	}
}

func (u *srv) Start() {
	userConfig := u.config.GetUser()
	if len(userConfig.ListenAddr) == 0 || userConfig.ListenPort <= 0 {
		panic(errors.New("config.listen_addr and config.listen_port must be positive"))
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", userConfig.ListenAddr, userConfig.ListenPort))
	if err != nil {
		panic(err)
	}
	protoValid, err := protovalidate.New()
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			protovalidatemiddleware.UnaryServerInterceptor(protoValid), // protoValidate验证参数
			utils.UnaryLogIDInterceptor(u.config.LogConfig),            // 将调用方传递的logId写入ctx中
			utils.UnaryLoggingInterceptor(u.config.LogConfig),          // log level为debug时记录请求及响应的日志
		),
	)
	userv1.RegisterUserServiceServer(s, u.user)
	u.grpServer = s
	if err = s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func (u *srv) Stop() {
	u.grpServer.GracefulStop()
	log.Println(logx.Sync())
}
