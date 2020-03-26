package synchronise

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/GroM1124/filemanager/engine"
	"github.com/GroM1124/filemanager/readdir"
)

type Synchronise struct {
	Path1 string `json:"path1"`
	Path2 string `json:"path2"`
	Mask  struct {
		On      bool     `json:"on"`
		Ext     []string `json:"ext"`
		Include bool     `json:"Include"`
		Verbose bool     `json:"verbose"`
	} `json:"mask"`
	GetHash bool `json:"getHash"`
}

func Run(config string) error {
	data, err := ioutil.ReadFile(config)
	if os.IsNotExist(err) {
		return fmt.Errorf("no config file")
	} else if err != nil {
		return fmt.Errorf("could not read config %v", err)
	}
	var s []Synchronise
	err = json.Unmarshal(data, &s)
	if err != nil {
		return fmt.Errorf("could not unmarshal config %v", err)
	}
	for _, sync := range s {
		if !sync.Mask.On {
			sync.Mask.Ext = nil
			sync.Mask.Include = false
			sync.Mask.Verbose = false
		}
		err := sync.Sync()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Synchronise) Sync() error {
	rd1 := readdir.SetRD(s.Path1, s.Mask.Ext, s.GetHash)
	rd2 := readdir.SetRD(s.Path2, s.Mask.Ext, s.GetHash)
	ex1, in1, err := rd1.ReadDir()
	if err != nil {
		return err
	}
	ex2, in2, err := rd2.ReadDir()
	if err != nil {
		return err
	}
	if s.Mask.Include {
		ex1, in1 = in1, ex1
		ex2, in2 = in2, ex2
	}
	if s.Mask.Verbose {
		excluded := append(ex1, ex2...)
		sort.Slice(excluded, func(i, j int) bool {
			return excluded[i].PathAbs < excluded[j].PathAbs
		})
		for _, fi := range excluded {
			fmt.Printf("%q\t%v\t%q\t%q\n", fi.PathAbs, fi.Size, fi.ModTime, "the file is excluded by a mask")
		}
	}
	arr, err := readResult()
	if err != nil {
		fmt.Printf("The first synchronization will take place\n\n")
	}
	for _, fi := range arr {
		fmt.Printf("%q\t%v\t%q\t%q\n", fi.PathAbs, fi.Size, fi.ModTime, "read")
	}
	res1 := engine.CompareSync(arr, in1, s.Path2)
	res2 := engine.CompareSync(arr, in2, s.Path1)
	match, dfr := engine.CompareResolution(res1, res2)
	for _, action := range match {
		fmt.Println(action)
	}
	if len(dfr) == 0 {
		fmt.Printf("No files for synchronization\n\n")
		return nil
	}
	for _, action := range dfr {
		fmt.Println(action)
	}
	fmt.Printf("Pleace enter \"Y\" for synchronization " +
		"or enter any other character to cancel synchronization\n\n")
	var ask string
	fmt.Scanln(&ask)
	if ask == "y" || ask == "Y" {
		for _, action := range dfr {
			err := action.Apply()
			if err != nil {
				return err
			}
		}
		fmt.Printf("Synchronize is done\n\n")
		err = s.writeResult(*rd1)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("Synchronize canceled by user\n\n")
	}
	return nil
}

func readResult() ([]readdir.FI, error) {
	var arr []readdir.FI
	r, err := os.Open("result.json")
	if err != nil {
		return nil, fmt.Errorf("could not open file result.json: %v", err)
	}
	defer r.Close()
	dec := json.NewDecoder(r)
	err = dec.Decode(&arr)
	if err != nil {
		return nil, err
	}
	return arr, nil
}

func (s *Synchronise) writeResult(rd readdir.RD) error {
	fmt.Println("Write result file.")
	ex, in, err := rd.ReadDir()
	if err != nil {
		return err
	}
	if s.Mask.Include {
		ex, in = in, ex
	}
	w, err := os.Create("result.json")
	if err != nil {
		return fmt.Errorf("could not create file result.json: %v", err)
	}
	defer w.Close()
	err = json.NewEncoder(w).Encode(in)
	if err != nil {
		return fmt.Errorf("could not encode file result.json: %v", err)
	}
	fmt.Println("Done.")
	return nil
}
