package main

import (
	"github.com/gromey/filemanager/reject"
	"log"
)

func main() {
	err := reject.Run("cmd/reject/config.json")
	if err != nil {
		log.Fatal(err)
	}
}
