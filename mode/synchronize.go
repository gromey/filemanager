package mode

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"syncdata/engine"
)

var path1 string
var path2 string
var ext []string

func Sync(c Config) error {
	path1 = c.Paths[0]
	path2 = c.Paths[1]
	if c.Mask.On {
		ext = c.Mask.Ext
	}
	ex1, in1, err := engine.ReadDir(path1, ext)
	if err != nil {
		return err
	}
	ex2, in2, err := engine.ReadDir(path2, ext)
	if err != nil {
		return err
	}
	include := false
	if c.Mask.On && c.Mask.Include {
		ex1, in1 = in1, ex1
		ex2, in2 = in2, ex2
		include = true
	}
	if c.Mask.On && c.Mask.Verbose {
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
	res1 := engine.CompareSync(arr, in1, path2)
	res2 := engine.CompareSync(arr, in2, path1)
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
		err = writeResult(include)
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

func writeResult(include bool) error {
	fmt.Println("Write result file.")
	ex, in, err := engine.ReadDir(path1, ext)
	if err != nil {
		return err
	}
	if include {
		ex, in = in, ex
	}
	w, err := os.Create("result.json")
	if err != nil {
		return fmt.Errorf("could not create file result.json: %v", err)
	}
	defer w.Close()
	json.NewEncoder(w).Encode(in)
	fmt.Println("Done.")
	return nil
}
