package synchronizer

import (
	"encoding/json"
	"fmt"
	"github.com/gromey/filemanager/dirreader"
	"os"
	"sort"
)

// synchronizer ...
type synchronizer struct {
	firstPath  string
	secondPath string
	ext        []string
	include    bool
	details    bool
	getHash    bool
}

// New ...
func New(c *Config) *synchronizer {
	sync := new(synchronizer)

	sync.firstPath = c.FirstPath
	sync.secondPath = c.SecondPath
	if c.Mask.On {
		sync.ext = c.Mask.Ext
		sync.include = c.Mask.Include
		sync.details = c.Mask.Details
	}
	sync.getHash = c.GetHash

	return sync
}

func (s *synchronizer) Start() error {
	ex1, in1, err := dirreader.SetDirReader(s.firstPath, s.ext, s.getHash).Exec()
	if err != nil {
		return err
	}

	ex2, in2, err := dirreader.SetDirReader(s.secondPath, s.ext, s.getHash).Exec()
	if err != nil {
		return err
	}

	if s.include {
		ex1, in1 = in1, ex1
		ex2, in2 = in2, ex2
	}

	if s.details {
		excluded := append(ex1, ex2...)
		sort.Slice(excluded, func(i, j int) bool {
			return excluded[i].PathAbs < excluded[j].PathAbs
		})

		//for _, fi := range excluded {
		//	fmt.Printf("%q\t%v\t%q\t%q\n", fi.PathAbs, fi.Size(), fi.ModTime(), "the file is excluded by a mask")
		//}
	}

	arr, err := readResult()
	if err != nil {
		fmt.Printf("The first synchronization will take place\n\n")
	}

	//for _, fi := range arr {
	//	fmt.Printf("%q\t%v\t%q\t%q\n", fi.PathAbs, fi.Size(), fi.ModTime(), "read")
	//}

	res1 := CompareSync(arr, in1, s.secondPath)
	res2 := CompareSync(arr, in2, s.firstPath)

	match, dfr := CompareResolution(res1, res2)
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
		return nil, fmt.Errorf("could not open file result.json: %v", err)
	}
	defer r.Close()

	err = json.NewDecoder(r).Decode(&arr)
	if err != nil {
		return nil, err
	}
	return arr, nil
}

func (s *synchronizer) writeResult(rd dirreader.DirReader) error {
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
