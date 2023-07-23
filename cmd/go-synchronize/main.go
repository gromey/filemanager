package main

import (
	"log"
)

func main() {
	if err := startInTerminal(); err != nil {
		log.Fatal(err)
	}
}
