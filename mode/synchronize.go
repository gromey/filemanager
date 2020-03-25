package mode

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/GroM1124/sync/engine"
)

type Syncer struct {
	path1   string
	path2   string
	mask    bool
	ext     []string
	include bool
	verbose bool
	getHash bool
}

func (s *Syncer) Sync() error {
	rd1 := engine.SetRD(s.path1, s.ext, s.getHash)
	rd2 := engine.SetRD(s.path2, s.ext, s.getHash)
	ex1, in1, err := rd1.ReadDir()
	if err != nil {
		return err
	}
	ex2, in2, err := rd2.ReadDir()
	if err != nil {
		return err
	}
	if s.include {
		ex1, in1 = in1, ex1
		ex2, in2 = in2, ex2
	}
	if s.verbose {
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
	res1 := engine.CompareSync(arr, in1, s.path2)
	res2 := engine.CompareSync(arr, in2, s.path1)
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

func readResult() ([]engine.FI, error) {
	var arr []engine.FI
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

func (s *Syncer) writeResult(rd engine.RD) error {
	fmt.Println("Write result file.")
	ex, in, err := rd.ReadDir()
	if err != nil {
		return err
	}
	if s.include {
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
