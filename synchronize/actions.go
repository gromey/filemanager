package synchronize

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gromey/filemanager/dirreader"
)

type applier interface {
	apply() error
}

type action int

const (
	actMatch = action(iota) + 1
	actCreate
	actReplace
	actRemove
	actQuestion
)

func (a action) String() string {
	switch a {
	case actMatch:
		return "match"
	case actCreate:
		return "create"
	case actReplace:
		return "replace"
	case actRemove:
		return "remove"
	case actQuestion:
		return "question"
	}
	return "unknown action"
}

type base struct {
	fiSrc dirreader.FileInfo
	fiDst dirreader.FileInfo
}

type match struct {
	*base
}

func (m *match) apply() error {
	return nil
}

func (m *match) String() string {
	return fmt.Sprintf("%q\tmatch\t%q\n\tSize %d\n\tModTime %s\n\tHash %q",
		m.fiSrc.PathAbs, m.fiDst.PathAbs, m.fiDst.Size,
		m.fiDst.ModTime, m.fiDst.Hash)
}

type create struct {
	*base
}

func (c *create) apply() error {
	src, err := os.Open(c.fiSrc.PathAbs)
	if err != nil {
		return fmt.Errorf("open file %s: %w", c.fiSrc.PathAbs, err)
	}
	defer func() {
		e := src.Close()
		if e != nil {
			fmt.Println(e)
		}
	}()

	path := filepath.Dir(c.fiDst.PathAbs)
	if err = os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("create Dir %s: %w", path, err)
	}

	var dst *os.File
	if dst, err = os.Create(c.fiDst.PathAbs); err != nil {
		return fmt.Errorf("create file %s: %w", c.fiDst.PathAbs, err)
	}

	defer func() {
		e := dst.Close()
		if e != nil {
			fmt.Println(e)
		}
	}()

	if _, err = io.Copy(dst, src); err != nil {
		_ = os.Remove(c.fiDst.PathAbs)
		return fmt.Errorf("copy file %s: %w", c.fiDst.PathAbs, err)
	}

	if err = os.Chtimes(c.fiDst.PathAbs, c.fiSrc.ModTime, c.fiSrc.ModTime); err != nil {
		return fmt.Errorf("change mod time %s: %w", c.fiDst.PathAbs, err)
	}

	return nil
}

func (c *create) String() string {
	return fmt.Sprintf("%q\t%d\t%s\tthe file will be creating in\t%s",
		c.fiSrc.Name, c.fiSrc.Size, c.fiSrc.ModTime, c.fiDst.PathAbs)
}

type replace struct {
	*base
}

func (r *replace) apply() error {
	src, err := os.Open(r.fiSrc.PathAbs)
	if err != nil {
		return fmt.Errorf("open file %s: %w", r.fiSrc.PathAbs, err)
	}
	defer func() { _ = src.Close() }()

	var dst *os.File
	if dst, err = os.Create(r.fiDst.PathAbs); err != nil {
		return fmt.Errorf("create file %s: %w", r.fiDst.PathAbs, err)
	}
	defer func() { _ = dst.Close() }()

	if _, err = io.Copy(dst, src); err != nil {
		_ = os.Remove(r.fiDst.PathAbs)
		return fmt.Errorf("copy file %s: %w", r.fiDst.PathAbs, err)
	}

	if err = os.Chtimes(r.fiDst.PathAbs, r.fiSrc.ModTime, r.fiSrc.ModTime); err != nil {
		return fmt.Errorf("change mod time %s: %w", r.fiDst.PathAbs, err)
	}

	return nil
}

func (r *replace) String() string {
	return fmt.Sprintf("%q\t%d\t%s\tthe file will be replacing in\t%s",
		r.fiSrc.Name, r.fiSrc.Size, r.fiSrc.ModTime, r.fiDst.PathAbs)
}

type remove struct {
	*base
}

func (r *remove) apply() error {
	if err := os.Remove(r.fiDst.PathAbs); err != nil {
		return fmt.Errorf("remove file %s: %w", r.fiDst.PathAbs, err)
	}
	return nil
}

func (r *remove) String() string {
	return fmt.Sprintf("%q\t%d\t%s\tthe file will be removed\t%s",
		r.fiDst.Name, r.fiDst.Size, r.fiDst.ModTime, r.fiDst.PathAbs)
}

type question struct {
	*base
}

func (q *question) apply() error {
	actA, actB := actReplace, actReplace

	switch {
	case q.fiSrc.Name != "" && q.fiDst.Name == "":
		actA, actB = actCreate, actRemove
	case q.fiSrc.Name == "" && q.fiDst.Name != "":
		actA, actB = actRemove, actCreate
	}

	fmt.Printf("Enter '>' for %s file in %q or enter '<' for %s file in %q %s",
		actA.String(), q.fiDst.PathAbs, actB.String(), q.fiSrc.PathAbs,
		"or enter any other character to skip the file synchronization\n")

	var ask string
	if _, err := fmt.Scanln(&ask); err != nil {
		return err
	}

	res := &resolution{base: q.base}

	if ask == ">" {
		res.action = actA
		return res.Apply()
	} else if ask == "<" {
		res.action = actB
		res.fiSrc, res.fiDst = res.fiDst, res.fiSrc
		return res.Apply()
	}

	fmt.Println("skipped")

	return nil
}
