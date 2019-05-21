package main

import (
	"log"

	"github.com/GroM1124/sync/mode"
)

func main() {
	err := mode.Run("config.json")
	if err != nil {
		log.Fatal(err)
	}
}
