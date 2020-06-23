package main

import (
	"log"

	"github.com/GroM1124/filemanager/synchronise"
)

func main() {
	err := synchronise.Run("cmd/go-synchronise/config.json")
	if err != nil {
		log.Fatal(err)
	}
}
