package dirreader

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileInfo struct {
	os.FileInfo
	PathAbs string
	PathRel string
	Hash    string
}

type dirReader struct {
	root    string
	mask    []string
	include bool
	details bool
	getHash bool
}

func SetDirReader(root string, mask []string, include, details, getHash bool) *dirReader {
	return &dirReader{
		root:    root,
		mask:    mask,
		include: include,
		details: details,
		getHash: getHash,
	}
}

func (r *dirReader) Exec() ([]FileInfo, []FileInfo, error) {
	ex, in, err := readDirectory(r.root, "", r.mask, r.getHash)
	if err != nil {
		return nil, nil, err
	}

	if r.include {
		ex, in = in, ex
	}

	if r.details {
		sort.Slice(ex, func(i, j int) bool {
			return ex[i].PathAbs < ex[j].PathAbs
		})
	}

	return ex, in, nil
}

func readDirectory(root, rel string, mask []string, getHash bool) ([]FileInfo, []FileInfo, error) {
	dir, err := os.Open(root)
	if err != nil {
		return nil, nil, fmt.Errorf("can't open %v: %v", root, err)
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, nil, fmt.Errorf("can't read files in dir %v: %v", root, err)
	}

	var inMask, outMask []FileInfo
	var abs, resultingHash string

	for _, file := range files {
		abs = filepath.Join(root, file.Name())

		if file.IsDir() {
			inM, outM, err := readDirectory(abs, filepath.Join(rel, file.Name()), mask, getHash)
			if err != nil {
				return nil, nil, err
			}

			inMask = append(inMask, inM...)
			outMask = append(outMask, outM...)

			continue
		}

		if getHash {
			resultingHash, err = computingHash(abs)
			if err != nil {
				return nil, nil, fmt.Errorf("can't compute hash for %v: %v", abs, err)
			}
		}

		if includedInMask(file.Name(), mask) {
			inMask = append(inMask, FileInfo{
				FileInfo: file,
				PathAbs:  abs,
				PathRel:  rel,
				Hash:     resultingHash,
			})
			continue
		} else {
			outMask = append(outMask, FileInfo{
				FileInfo: file,
				PathAbs:  abs,
				PathRel:  rel,
				Hash:     resultingHash,
			})
		}
	}
	return inMask, outMask, nil
}

func computingHash(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func includedInMask(name string, mask []string) bool {
	for _, ext := range mask {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}
