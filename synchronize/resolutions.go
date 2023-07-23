package synchronize

import (
	"fmt"
	"path/filepath"

	"github.com/gromey/filemanager/dirreader"
)

type Resolution interface {
	Apply() error
	String() string
}

type Resolutions map[string]Resolution

func makeResolution(action action, fiSrc, fiDst dirreader.FileInfo) *resolution {
	return &resolution{
		action: action,
		base:   &base{fiSrc: fiSrc, fiDst: fiDst},
	}
}

func makeResolutionMatch(fiSrc, fiDst dirreader.FileInfo) *resolution {
	return makeResolution(actMatch, fiSrc, fiDst)
}

func makeResolutionCreate(fiSrc dirreader.FileInfo, abs, relName string) *resolution {
	return makeResolution(actCreate, fiSrc, dirreader.FileInfo{PathAbs: filepath.Join(abs, relName)})
}

func makeResolutionReplace(fiSrc, fiDst dirreader.FileInfo) *resolution {
	return makeResolution(actReplace, fiSrc, fiDst)
}

func makeResolutionRemove(fiSrc dirreader.FileInfo, abs, relName string) *resolution {
	return makeResolution(actRemove, fiSrc, dirreader.FileInfo{PathAbs: filepath.Join(abs, relName)})
}

func makeResolutionQuestion(fiSrc, fiDst dirreader.FileInfo) *resolution {
	return makeResolution(actQuestion, fiSrc, fiDst)
}

type resolution struct {
	action action
	*base
}

func (r *resolution) Apply() error {
	var a applier
	switch r.action {
	case actMatch:
		a = &match{base: r.base}
	case actCreate:
		a = &create{base: r.base}
	case actReplace:
		a = &replace{base: r.base}
	case actRemove:
		a = &remove{base: r.base}
	case actQuestion:
		a = &question{base: r.base}
	default:
		return nil
	}
	return a.apply()
}

func (r *resolution) String() string {
	var s string
	switch r.action {
	case actMatch:
		s = fmt.Sprintf("%s has not been changed", r.fiSrc.PathAbs)
	case actCreate:
		s = fmt.Sprintf("%s was created, time %q", r.fiDst.PathAbs, r.fiSrc.ModTime)
	case actReplace:
		s = fmt.Sprintf("%s has been changed %q", r.fiDst.PathAbs, r.fiSrc.ModTime)
	case actRemove:
		s = fmt.Sprintf("%s has been deleted", r.fiDst.PathAbs)
	case actQuestion:
		s = fmt.Sprintf("%s conflicts with %s", r.fiSrc.PathAbs, r.fiDst.PathAbs)
	default:
		return "Unknown resolution"
	}
	return s
}
