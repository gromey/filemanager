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
	"time"
)

type DirReader interface {
	Exec() (excl []FileInfo, incl []FileInfo, err error)
}

type FileInfo struct {
	Name    string
	Size    int64
	ModTime time.Time
	IsDir   bool
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

func New(root string, mask []string, include, getHash bool) DirReader {
	return &dirReader{
		root:    root,
		mask:    mask,
		include: include,
		getHash: getHash,
	}
}

func (r *dirReader) Exec() (excl []FileInfo, incl []FileInfo, err error) {
	if excl, incl, err = readDirectory(r.root, "", r.mask, r.getHash); err != nil {
		return
	}

	if r.include {
		excl, incl = incl, excl
	}

	sort.Slice(incl, func(i, j int) bool {
		return incl[i].PathAbs < incl[j].PathAbs
	})

	return
}

func readDirectory(root, rel string, mask []string, getHash bool) ([]FileInfo, []FileInfo, error) {
	dir, err := os.Open(root)
	if err != nil {
		return nil, nil, fmt.Errorf("read directory: %w", err)
	}
	defer func() { _ = dir.Close() }()

	var files []os.FileInfo
	if files, err = dir.Readdir(-1); err != nil {
		return nil, nil, fmt.Errorf("can't read files in dir %s: %w", root, err)
	}

	var inMask, outMask []FileInfo
	var abs, resultingHash string

	for _, file := range files {
		abs = filepath.Join(root, file.Name())

		if file.IsDir() {
			var inM, outM []FileInfo
			if inM, outM, err = readDirectory(abs, filepath.Join(rel, file.Name()), mask, getHash); err != nil {
				return nil, nil, err
			}

			inMask = append(inMask, inM...)
			outMask = append(outMask, outM...)

			continue
		}

		if getHash {
			if resultingHash, err = computingHash(abs); err != nil {
				return nil, nil, fmt.Errorf("can't compute hash for %s: %w", abs, err)
			}
		}

		fi := FileInfo{
			Name:    file.Name(),
			Size:    file.Size(),
			ModTime: file.ModTime(),
			IsDir:   file.IsDir(),
			PathAbs: abs,
			PathRel: rel,
			Hash:    resultingHash,
		}

		if includedInMask(file.Name(), mask) {
			inMask = append(inMask, fi)
		} else {
			outMask = append(outMask, fi)
		}
	}

	return inMask, outMask, nil
}

func computingHash(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	h := md5.New()
	if _, err = io.Copy(h, file); err != nil {
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
