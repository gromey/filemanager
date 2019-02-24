package mode

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	Mode  Mode     `json:"mode"`
	Paths []string `json:"paths"`
	Mask  struct {
		On      bool     `json:"on"`
		Ext     []string `json:"ext"`
		Include bool     `json:"Include"`
		Verbose bool     `json:"verbose"`
	} `json:"mask"`
}

type Mode int

const (
	Synchronize = Mode(iota + 1)
	Duplicate
)

func Run(config string) error {
	data, err := ioutil.ReadFile(config)
	if os.IsNotExist(err) {
		return fmt.Errorf("no config file")
	} else if err != nil {
		return fmt.Errorf("could not read config %v", err)
	}
	var c []Config
	err = json.Unmarshal(data, &c)
	if err != nil {
		return fmt.Errorf("could not unmarshal config %v", err)
	}
	for _, config := range c {
		switch config.Mode {
		case Synchronize:
			err := Sync(config)
			if err != nil {
				return err
			}
		case Duplicate:
			err := Dupl(config)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
