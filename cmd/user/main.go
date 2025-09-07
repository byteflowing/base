package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/byteflowing/base/pkg/user"
	"github.com/byteflowing/go-common/signalx"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	userv1 "github.com/byteflowing/proto/gen/go/services/user/v1"
)

func main() {
	configPath := flag.String("config", "config.db.yaml", "path to config file")
	flag.Parse()
	sigListener := signalx.NewSignalListener()
	userImpl := NewWithConfig(*configPath)
	userService := &UserService{
		config: userImpl.GetConfig(),
		user:   userImpl,
	}
	sigListener.Register(userService)
	sigListener.Listen()
	log.Printf("exit")
}

type UserService struct {
	user      *user.Impl
	config    *configv1.Config
	grpServer *grpc.Server
}

func (u *UserService) Start() {
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

func (u *UserService) Stop() {
	u.grpServer.GracefulStop()
}
