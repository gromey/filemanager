package synchronize

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gromey/filemanager/dirreader"
)

type Synchronize interface {
	Start() (res1, res2 map[string]*resolution, msg string, err error)
}

type Config struct {
	FirstPath  string `json:"first_path"`
	SecondPath string `json:"second_path"`
	Mask       struct {
		On        bool     `json:"on"`
		Extension []string `json:"extension"`
		Include   bool     `json:"include"`
		Details   bool     `json:"details"`
	} `json:"mask"`
	GetHash bool `json:"get_hash"`
}

type synchronize struct {
	firstPath  string
	secondPath string
	ext        []string
	include    bool
	details    bool
	getHash    bool
}

func New(c *Config) Synchronize {
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

func (s *synchronize) Start() (res1, res2 map[string]*resolution, msg string, err error) {
	var excluded1, included1, excluded2, included2, res []dirreader.FileInfo

	if excluded1, included1, err = dirreader.New(s.firstPath, s.ext, s.include, s.details, s.getHash).Exec(); err != nil {
		return
	}

	if excluded2, included2, err = dirreader.New(s.secondPath, s.ext, s.include, s.details, s.getHash).Exec(); err != nil {
		return
	}

	if s.details {
		log.Printf("%d files was excluded by mask.\n", len(excluded1)+len(excluded2))
	}

	if res, err = readPreviousResult("result.json"); err != nil && !os.IsNotExist(err) {
		return
	} else {
		msg = "The first synchronization will take place.\n\n"
	}

	res1 = compareFileInfo(res, included1, s.secondPath)
	res2 = compareFileInfo(res, included2, s.firstPath)

	return
}

func readPreviousResult(filename string) ([]dirreader.FileInfo, error) {
	var res []dirreader.FileInfo

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("can't unmarshal %s: %s", filename, err)
	}

	return res, nil
}

//	match, dfr := CompareResolution(res1, res2)
//	// return match, dfr, err
//
//	//	// TODO Move to other func
//	//	for _, action := range match {
//	//		fmt.Println(action)
//	//	}
//	//	if len(dfr) == 0 {
//	//		fmt.Printf("No files for synchronization\n\n")
//	//		return nil
//	//	}
//	//	for _, action := range dfr {
//	//		fmt.Println(action)
//	//	}
//	//	fmt.Printf("Please enter \"Y\" for synchronization " +
//	//		"or enter any other character to cancel synchronization\n\n")
//	//	var ask string
//	//	fmt.Scanln(&ask)
//	//	if ask == "y" || ask == "Y" {
//	//		for _, action := range dfr {
//	//			err := action.Apply()
//	//			if err != nil {
//	//				return err
//	//			}
//	//		}
//	//		fmt.Printf("Synchronize is done\n\n")
//	//
//	//		err = s.writeResult(*dir1)
//	//		if err != nil {
//	//			return err
//	//		}
//	//	} else {
//	//		fmt.Printf("Synchronize canceled by user\n\n")
//	//	}
//	return nil
//}

//// TODO Refactor it func
//func (s *synchronize) writeResult(rd dirreader.DirReader) error {
//	fmt.Println("Write result file.")
//	ex, in, err := rd.ReadDirectory()
//	if err != nil {
//		return err
//	}
//	if s.config.Mask.include {
//		ex, in = in, ex
//	}
//	w, err := os.Create("result.json")
//	if err != nil {
//		return fmt.Errorf("could not create file result.json: %v", err)
//	}
//	defer w.Close()
//	err = json.NewEncoder(w).Encode(in)
//	if err != nil {
//		return fmt.Errorf("could not encode file result.json: %v", err)
//	}
//	fmt.Println("Done.")
//	return nil
//}
