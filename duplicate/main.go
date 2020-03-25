package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"github.com/GroM1124/filemanager/engine"
	"github.com/GroM1124/filemanager/readdir"
)

func main() {
	err := Run("duplicate/config.json")
	if err != nil {
		log.Fatal(err)
	}
}

type Config struct {
	Paths []string `json:"paths"`
	Mask  struct {
		On      bool     `json:"on"`
		Ext     []string `json:"ext"`
		Include bool     `json:"Include"`
		Verbose bool     `json:"verbose"`
	} `json:"mask"`
}

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
		err := Dupl(config)
		if err != nil {
			return err
		}
	}
	return nil
}

func Dupl(c Config) error {
	var ext []string
	if c.Mask.On {
		ext = c.Mask.Ext
	}
	var excl, incl []readdir.FI
	for _, path := range c.Paths {
		rd := readdir.SetRD(path, ext, true)
		ex, in, err := rd.ReadDir()
		if err != nil {
			return err
		}
		excl = append(excl, ex...)
		incl = append(incl, in...)
	}
	if c.Mask.On && c.Mask.Include {
		excl, incl = incl, excl
	}
	if c.Mask.On && c.Mask.Verbose {
		sort.Slice(excl, func(i, j int) bool {
			return excl[i].PathAbs < excl[j].PathAbs
		})
		for _, fi := range excl {
			fmt.Printf("%q\t%v\t%q\t%q\n", fi.PathAbs, fi.Size, fi.ModTime, "the file is excluded by a mask")
		}
	}
	match := engine.CompareDpl(incl)
	if len(match) == 0 {
		fmt.Printf("No match\n\n")
		return nil
	}
	for _, action := range match {
		fmt.Println(action)
	}
	return nil
}
