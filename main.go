package main

import (
	"log"

	"github.com/GroM1124/filemanager/duplicate"
	"github.com/GroM1124/filemanager/synchronise"
)

func main() {
	var err error
	sync := true
	dupl := true
	if sync {
		err = synchronise.Run("synchronise/config.json")
		if err != nil {
			log.Fatal(err)
		}
	}
	if dupl {
		err = duplicate.Run("duplicate/config.json")
		if err != nil {
			log.Fatal(err)
		}
	}
}
