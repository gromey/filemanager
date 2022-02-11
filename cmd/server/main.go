package main

import (
	"github.com/gromey/filemanager/server"
	"log"
)

func main() {
	cfg := server.NewConfig()

	s := server.New(cfg)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
