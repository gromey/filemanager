package main

import (
	"github.com/GroM1124/filemanager/reject"
	"log"
)

func main() {
	err := reject.Run("cmd/reject/config.json")
	if err != nil {
		log.Fatal(err)
	}
}
