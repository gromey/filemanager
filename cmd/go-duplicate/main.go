package main

import (
	"log"
	"time"

	"github.com/gromey/filemanager/common"
	"github.com/gromey/filemanager/duplicate"
)

func main() {
	start := time.Now()
	if err := run("cmd/go-duplicate/config.json"); err != nil {
		log.Fatal(err)
	}
	log.Printf("%s", time.Since(start))
}

func run(config string) error {
	var c []*duplicate.Config

	if err := common.ReadConfig(config, &c); err != nil {
		return err
	}

	for _, cfg := range c {
		res, err := duplicate.New(cfg).Start()
		if err != nil {
			return err
		}

		for _, v := range res {
			log.Println(v)
		}
	}

	return nil
}
