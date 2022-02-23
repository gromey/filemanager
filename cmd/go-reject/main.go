package main

import (
	"github.com/gromey/filemanager/reject"
	"log"
	"time"
)

func main() {
	start := time.Now()
	if err := reject.Run("cmd/go-reject/config.json"); err != nil {
		log.Fatal(err)
	}
	log.Printf("%s", time.Since(start))
}
