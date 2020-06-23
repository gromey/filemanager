package main

import (
	"log"

	"github.com/GroM1124/filemanager/duplicate"
)

func main() {
	err := duplicate.Run("cmd/go-duplicate/config.json")
	if err != nil {
		log.Fatal(err)
	}
}
