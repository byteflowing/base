package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/byteflowing/base/pkg/user"
	"github.com/byteflowing/go-common/logx"
	"github.com/byteflowing/go-common/signalx"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	userv1 "github.com/byteflowing/proto/gen/go/services/user/v1"
)

func main() {
	configPath := flag.String("config", "./config.dev.yaml", "user -config config.yaml")
	flag.Parse()
	userImpl := NewWithConfig(*configPath)
	config := userImpl.GetConfig()
	logx.Init(config.LogConfig)
	userService := newSrv(userImpl, config)
	signalListener := signalx.NewSignalListener(30 * time.Second)
	signalListener.Register(userService)
	signalListener.Listen()
	log.Println("user service exited")
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
	s := grpc.NewServer()
	userv1.RegisterUserServiceServer(s, u.user)
	userConfig := u.config.GetUser()
	if len(userConfig.ListenAddr) == 0 || userConfig.ListenPort <= 0 {
		panic(errors.New("config.listen_addr and config.listen_port must be positive"))
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", userConfig.ListenAddr, userConfig.ListenPort))
	if err != nil {
		panic(err)
	}
	u.grpServer = s
	if err = s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func (u *srv) Stop() {
	u.grpServer.GracefulStop()
	log.Println(logx.Sync())
}
