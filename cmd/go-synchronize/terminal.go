package main

import (
	"fmt"

	"github.com/gromey/filemanager/common"
	"github.com/gromey/filemanager/synchronize"
)

func startInTerminal() error {
	cfg := new(synchronize.Config)
	if err := common.ReadConfig("cmd/go-synchronize/config.json", cfg); err != nil {
		return err
	}

	s := synchronize.New(cfg)

	_, diff, err := s.Start()
	if err != nil {
		return err
	}

	if len(diff) == 0 {
		fmt.Printf("No files for synchronization\n")
		return nil
	}

	for _, res := range diff {
		fmt.Println(res)
	}

	fmt.Printf("Please enter \"Y\" for synchronization " +
		"or enter any other character to cancel synchronization\n")

	var ask string
	if _, err = fmt.Scanln(&ask); err != nil {
		return err
	}

	if ask == "y" || ask == "Y" {
		for _, res := range diff {
			if err = res.Apply(); err != nil {
				return err
			}
		}
		fmt.Printf("Synchronize is done\n")
	} else {
		fmt.Printf("Synchronize canceled by user\n")
	}

	return nil
}
