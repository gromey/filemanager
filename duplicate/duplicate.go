package duplicate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/GroM1124/filemanager/engine"
	"github.com/GroM1124/filemanager/readdir"
)

type Duplicate struct {
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
	var d []Duplicate
	err = json.Unmarshal(data, &d)
	if err != nil {
		return fmt.Errorf("could not unmarshal config %v", err)
	}
	for _, duplicate := range d {
		err := duplicate.Dupl()
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Duplicate) Dupl() error {
	var ext []string
	if d.Mask.On {
		ext = d.Mask.Ext
	}
	var excl, incl []readdir.FI
	for _, path := range d.Paths {
		rd := readdir.SetRD(path, ext, true)
		ex, in, err := rd.ReadDir()
		if err != nil {
			return err
		}
		excl = append(excl, ex...)
		incl = append(incl, in...)
	}
	if d.Mask.On && d.Mask.Include {
		excl, incl = incl, excl
	}
	if d.Mask.On && d.Mask.Verbose {
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
