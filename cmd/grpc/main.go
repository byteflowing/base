package main

import (
	base "github.com/byteflowing/base/kitex_gen/base/baseservice"
	"log"
)

func main() {
	svr := base.NewServer(new(BaseServiceImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
