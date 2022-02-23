package main

import (
	"github.com/gromey/filemanager/duplicate"
	"log"
	"time"
)

func main() {
	start := time.Now()
	if err := duplicate.Run("cmd/go-duplicate/config.json"); err != nil {
		log.Fatal(err)
	}
	log.Printf("%s", time.Since(start))
}
