package main

import (
	"github.com/GroM1124/filemanager/internal/apiserver"
	"log"
)

func main() {
	cfg := apiserver.NewConfig()

	s := apiserver.New(cfg)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
