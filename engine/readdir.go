package engine

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FI struct {
	IsDir   bool
	Size    int64
	ModTime time.Time
	Name    string
	PathAbs string
	PathRel string
	Hash    string
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
		return nil, nil, fmt.Errorf("could not read names in dir %v: %v", root, err)
	}
	var inMask []FI
	var outMask []FI
	for _, file := range files {
		abs := filepath.Join(root, file.Name())
		var hash string
		if !file.IsDir() {
			hash, err = getHash(abs)
			if err != nil {
				return nil, nil, fmt.Errorf("could not get hash for %v: %v", abs, err)
			}
		}
		if maskFilter(file.Name(), mask) {
			inMask = append(inMask, FI{
				IsDir:   file.IsDir(),
				Size:    file.Size(),
				ModTime: file.ModTime(),
				Name:    file.Name(),
				PathAbs: abs,
				PathRel: rel,
				Hash:    hash,
			})
			continue
		} else if !file.IsDir() {
			outMask = append(outMask, FI{
				IsDir:   file.IsDir(),
				Size:    file.Size(),
				ModTime: file.ModTime(),
				Name:    file.Name(),
				PathAbs: abs,
				PathRel: rel,
				Hash:    hash,
			})
		}
		if file.IsDir() {
			inM, outM, err := readDir(abs, filepath.Join(rel, file.Name()), mask)
			if err != nil {
				return nil, nil, err
			}
			inMask = append(inMask, inM...)
			outMask = append(outMask, outM...)
		}
	}
	return inMask, outMask, nil
}

func maskFilter(name string, mask []string) bool {
	for _, ext := range mask {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

func getHash(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
