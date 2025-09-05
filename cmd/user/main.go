package main

import (
	"log"

	"github.com/byteflowing/base/pkg/user"
	"github.com/byteflowing/go-common/signalx"
)

func main() {
	sigListener := signalx.NewSignalListener()
	userService := user.NewWithConfig("")
	sigListener.Register(userService)
	sigListener.Listen()
	log.Printf("exit")
}
