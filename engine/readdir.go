package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FI struct {
	Abs  string
	Rel  string
	Name string
	Size int64
	Time time.Time
	Dir  bool
}

func ReadDir(root string, mask []string) ([]FI, []FI, error) {
	return readDir(root, "", mask)
}

func readDir(root, rel string, mask []string) ([]FI, []FI, error) {
	dh, err := os.Open(root)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open %v: %v", root, err)
	}
	defer dh.Close()
	files, err := dh.Readdir(-1)
	if err != nil {
		return nil, nil, fmt.Errorf("could not read Dir names in %v: %v", root, err)
	}
	var insideMask []FI
	var outsideMask []FI
	for _, file := range files {
		if maskFilter(file.Name(), mask) == true {
			insideMask = append(insideMask, FI{
				Abs:  filepath.Join(root, file.Name()),
				Rel:  rel,
				Name: file.Name(),
				Size: file.Size(),
				Time: file.ModTime(),
				Dir:  file.IsDir(),
			})
			continue
		} else if !file.IsDir() {
			outsideMask = append(outsideMask, FI{
				Abs:  filepath.Join(root, file.Name()),
				Rel:  rel,
				Name: file.Name(),
				Size: file.Size(),
				Time: file.ModTime(),
				Dir:  file.IsDir(),
			})
		}
		if file.IsDir() {
			path := filepath.Join(root, file.Name())
			inM, noM, err := readDir(path, filepath.Join(rel, file.Name()), mask)
			if err != nil {
				return nil, nil, err
			}
			insideMask = append(insideMask, inM...)
			outsideMask = append(outsideMask, noM...)
		}
	}
	return insideMask, outsideMask, nil
}

func maskFilter(name string, mask []string) bool {
	for _, ext := range mask {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}
