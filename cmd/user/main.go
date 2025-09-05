package main

import (
	"flag"
	"log"

	"github.com/byteflowing/go-common/signalx"
)

func main() {
	configPath := flag.String("config", "config.db.yaml", "path to config file")
	flag.Parse()
	sigListener := signalx.NewSignalListener()
	userService := NewWithConfig(*configPath)
	sigListener.Register(userService)
	sigListener.Listen()
	log.Printf("exit")
}
