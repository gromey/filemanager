package main

import (
	"github.com/gromey/filemanager/duplicate"
	"log"
)

func main() {
	err := duplicate.Run("cmd/go-duplicate/config.json")
	if err != nil {
		log.Fatal(err)
	}
}
