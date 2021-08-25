package dirreader

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// dirReader ...
type dirReader struct {
	root    string
	mask    []string
	getHash bool
}

// SetDirReader ...
func SetDirReader(root string, mask []string, gH bool) *dirReader {
	return &dirReader{
		root:    root,
		mask:    mask,
		getHash: gH,
	}
}

// Exec ...
func (dr *dirReader) Exec() ([]FileInfo, []FileInfo, error) {
	return readDirectory(dr.root, "", dr.mask, dr.getHash)
}

// FileInfo ...
type FileInfo struct {
	os.FileInfo
	PathAbs string
	PathRel string
	Hash    string
}

// readDirectory ...
func readDirectory(root, rel string, mask []string, gH bool) ([]FileInfo, []FileInfo, error) {
	dir, err := os.Open(root)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open %v: %v", root, err)
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, nil, fmt.Errorf("could not read names in dir %v: %v", root, err)
	}

	var inMask, outMask []FileInfo
	var abs, hash string

	for _, file := range files {
		abs = filepath.Join(root, file.Name())
		if !file.IsDir() && gH {
			hash, err = getHash(abs)
			if err != nil {
				return nil, nil, fmt.Errorf("could not get hash for %v: %v", abs, err)
			}
		}
		if maskFilter(file.Name(), mask) {
			inMask = append(inMask, FileInfo{
				FileInfo: file,
				PathAbs:  abs,
				PathRel:  rel,
				Hash:     hash,
			})
			continue
		} else if !file.IsDir() {
			outMask = append(outMask, FileInfo{
				FileInfo: file,
				PathAbs:  abs,
				PathRel:  rel,
				Hash:     hash,
			})
		}

		if file.IsDir() {
			inM, outM, err := readDirectory(abs, filepath.Join(rel, file.Name()), mask, gH)
			if err != nil {
				return nil, nil, err
			}
			inMask = append(inMask, inM...)
			outMask = append(outMask, outM...)
		}
	}
	return inMask, outMask, nil
}

// getHash ...
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

// maskFilter ...
func maskFilter(name string, mask []string) bool {
	for _, ext := range mask {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}
