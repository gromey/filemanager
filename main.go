package main

import (
	"log"
	"syncdata/mode"
)

func main() {
	err := mode.Run("config.json")
	if err != nil {
		log.Fatal(err)
	}
}
