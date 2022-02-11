package synchronize

import (
	"encoding/json"
	"fmt"
	"github.com/gromey/filemanager/dirreader"
	"log"
	"os"
)

type synchronize struct {
	firstPath  string
	secondPath string
	ext        []string
	include    bool
	details    bool
	getHash    bool
}

func New(c *Config) *synchronize {
	s := &synchronize{
		firstPath:  c.FirstPath,
		secondPath: c.SecondPath,
		getHash:    c.GetHash,
	}

	if c.Mask.On {
		s.ext = c.Mask.Extension
		s.include = c.Mask.Include
		s.details = c.Mask.Details
	}

	return s
}

func (s *synchronize) Start() error {
	excluded1, included1, err := dirreader.SetDirReader(s.firstPath, s.ext, s.include, s.details, s.getHash).Exec()
	if err != nil {
		return err
	}

	excluded2, included2, err := dirreader.SetDirReader(s.secondPath, s.ext, s.include, s.details, s.getHash).Exec()
	if err != nil {
		return err
	}

	if s.details {
		log.Printf("%d %s\n", len(excluded1)+len(excluded2), "files was excluded by mask.")
	}

	arr, err := readResult()
	if err != nil {
		fmt.Printf("The first synchronization will take place\n\n")
	}

	//for _, fi := range arr {
	//	fmt.Printf("%q\t%v\t%q\t%q\n", fi.PathAbs, fi.Size(), fi.ModTime(), "read")
	//}

	res1 := compare(arr, included1, s.secondPath)
	res2 := compare(arr, included2, s.firstPath)

	match, dfr := CompareResolution(res1, res2)
	// return match, dfr, err

	// TODO Move to other func
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

		err = s.writeResult(*dir1)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("Synchronize canceled by user\n\n")
	}
	return nil
}

func readResult() ([]dirreader.FileInfo, error) {
	var arr []dirreader.FileInfo

	r, err := os.Open("result.json")
	if err != nil {
		return nil, fmt.Errorf("can't open file result.json: %v", err)
	}
	defer r.Close()

	err = json.NewDecoder(r).Decode(&arr)
	if err != nil {
		return nil, err
	}
	return arr, nil
}

// TODO Refactor it func
func (s *synchronize) writeResult(rd dirreader.DirReader) error {
	fmt.Println("Write result file.")
	ex, in, err := rd.ReadDirectory()
	if err != nil {
		return err
	}
	if s.config.Mask.include {
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
