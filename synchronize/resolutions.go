package synchronize

import (
	"fmt"
	"path/filepath"

	"github.com/gromey/filemanager/dirreader"
)

type resolution struct {
	action act
	base
}

func (r *resolution) String() string {
	var s string
	switch r.action {
	case actMatch:
		s = fmt.Sprintf("%s has not been changed", r.fiSrc.PathAbs)
	case actCreate:
		s = fmt.Sprintf("%s was created, time %q", r.fiSrc.PathAbs, r.fiSrc.ModTime())
	case actReplace:
		s = fmt.Sprintf("%s has been changed %q", r.fiSrc.PathAbs, r.fiSrc.ModTime())
	case actDelete:
		s = fmt.Sprintf("%s has been deleted", r.fiSrc.PathAbs)
	case actProblem:
		s = fmt.Sprintf("%s has time earlier than the previous synchronization", r.fiSrc.PathAbs)
	default:
		return "Unknown resolution"
	}
	return s
}

func (r *resolution) createAction() Action {
	var action Action
	switch r.action {
	case actMatch:
		action = &r.base
	case actCreate:
		action = &Create{r.base}
	case actReplace:
		action = &Replace{r.base}
	case actDelete:
		action = &Delete{r.base}
	}
	return action
}

func makeResolutionMatch(fiSrc, fiDst dirreader.FileInfo) *resolution {
	return &resolution{
		action: actMatch,
		base: base{
			fiSrc: fiSrc,
			fiDst: fiDst,
		},
	}
}

func makeResolutionCreate(fiSrc dirreader.FileInfo, abs, relName string) *resolution {
	return &resolution{
		action: actCreate,
		base: base{
			fiSrc: fiSrc,
			fiDst: dirreader.FileInfo{
				PathAbs: filepath.Join(abs, relName),
			},
		},
	}
}

func makeResolutionReplace(fiSrc dirreader.FileInfo, abs, relName string) *resolution {
	return &resolution{
		action: actReplace,
		base: base{
			fiSrc: fiSrc,
			fiDst: dirreader.FileInfo{
				PathAbs: filepath.Join(abs, relName),
			},
		},
	}
}

func makeResolutionDelete(fiSrc dirreader.FileInfo, abs, relName string) *resolution {
	return &resolution{
		action: actDelete,
		base: base{
			fiSrc: fiSrc,
			fiDst: dirreader.FileInfo{
				PathAbs: filepath.Join(abs, relName),
			},
		},
	}
}

func makeResolutionProblem(fiSrc dirreader.FileInfo, abs, relName string) *resolution {
	return &resolution{
		action: actProblem,
		base: base{
			fiSrc: fiSrc,
			fiDst: dirreader.FileInfo{
				PathAbs: filepath.Join(abs, relName),
			},
		},
	}
}
