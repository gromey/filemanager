package synchronize

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gromey/filemanager/dirreader"
)

type Synchronize interface {
	Start() (match, difference Resolutions, err error)
	WriteResult() error
}

type Config struct {
	PathA string `json:"path_A"`
	PathB string `json:"path_B"`
	Mask  struct {
		On         bool     `json:"on"`
		Extensions []string `json:"extensions"`
		Include    bool     `json:"include"`
		Details    bool     `json:"details"`
	} `json:"mask"`
	GetHash bool `json:"get_hash"`
}

type synchronize struct {
	pathA      string
	pathB      string
	extensions []string
	include    bool
	details    bool
	getHash    bool

	result string
}

func New(c *Config) Synchronize {
	s := &synchronize{
		pathA:   c.PathA,
		pathB:   c.PathB,
		getHash: c.GetHash,
	}

	if c.Mask.On {
		s.extensions = c.Mask.Extensions
		s.include = c.Mask.Include
		s.details = c.Mask.Details
	}

	return s
}

func (s *synchronize) Start() (match, difference Resolutions, err error) {
	if s.result, err = resultName(s.pathA, s.pathB); err != nil {
		return
	}

	var exclA, inclA, exclB, inclB, previous []dirreader.FileInfo

	if previous, err = readPreviousResult(s.result); err != nil && !os.IsNotExist(err) {
		return
	} else if os.IsNotExist(err) {
		//msg = "The first synchronization will take place.\n\n"
	}

	if exclA, inclA, err = dirreader.New(s.pathA, s.extensions, s.include, s.getHash).Exec(); err != nil {
		return
	}

	if exclB, inclB, err = dirreader.New(s.pathB, s.extensions, s.include, s.getHash).Exec(); err != nil {
		return
	}

	if s.details {
		log.Printf("%d files was excluded by mask.\n", len(exclA)+len(exclB))
	}

	match, difference = s.compareFileInfo(previous, inclA, inclB)

	return
}

func readPreviousResult(filename string) ([]dirreader.FileInfo, error) {
	var res []dirreader.FileInfo

	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(file, &res); err != nil {
		return nil, fmt.Errorf("can't unmarshal %s: %w", filename, err)
	}

	return res, nil
}

func (s *synchronize) WriteResult() error {
	fmt.Println("Write result file.")

	_, incl, err := dirreader.New(s.pathA, s.extensions, s.include, s.getHash).Exec()
	if err != nil {
		return err
	}

	path := filepath.Dir(s.result)
	if err = os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("could not create Dir %s: %w", path, err)
	}

	var file *os.File
	if file, err = os.Create(s.result); err != nil {
		return fmt.Errorf("could not create result file: %v", err)
	}
	defer func() { _ = file.Close() }()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "\t")
	if err = enc.Encode(incl); err != nil {
		return fmt.Errorf("could not encode result: %v", err)
	}

	fmt.Println("Done.")
	return nil
}

func resultName(pathA, pathB string) (string, error) {
	h := md5.New()
	if _, err := io.WriteString(h, pathA); err != nil {
		return "", fmt.Errorf("result name: %w", err)
	}
	if _, err := io.WriteString(h, pathB); err != nil {
		return "", fmt.Errorf("result name: %w", err)
	}
	return filepath.Join(".", "results", hex.EncodeToString(h.Sum(nil))), nil
}
