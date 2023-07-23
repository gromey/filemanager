package synchronize

import (
	"path/filepath"

	"github.com/gromey/filemanager/dirreader"
)

func toMap(arr []dirreader.FileInfo) map[string]dirreader.FileInfo {
	m := make(map[string]dirreader.FileInfo, len(arr))
	for _, fi := range arr {
		m[filepath.Join(fi.PathRel, fi.Name)] = fi
	}
	return m
}

func (s *synchronize) compareFileInfo(previous, srcA, srcB []dirreader.FileInfo) (Resolutions, Resolutions) {
	mP, mA, mB := toMap(previous), toMap(srcA), toMap(srcB)

	mth, dif := make(Resolutions), make(Resolutions)

	var prev dirreader.FileInfo

	for relName, fA := range mA {
		fB, ok := mB[relName]
		if !ok {
			if prev, ok = mP[relName]; !ok {
				dif[relName] = makeResolutionCreate(fA, s.pathB, relName)
				continue
			}

			dif[relName] = s.makeResolutionWithPrevious(prev, fA, fB, relName)
			continue
		}

		delete(mB, relName)

		if fA.ModTime.Equal(fB.ModTime) {
			mth[relName] = makeResolutionMatch(fA, fB)
			continue
		}

		if prev, ok = mP[relName]; !ok {
			dif[relName] = makeResolutionQuestion(fA, fB)
			continue
		}

		dif[relName] = s.makeResolutionWithPrevious(prev, fA, fB, relName)
	}

	for relName, fB := range mB {
		if fA, ok := mA[relName]; !ok {
			if prev, ok = mP[relName]; !ok {
				dif[relName] = makeResolutionCreate(fB, s.pathA, relName)
				continue
			}

			dif[relName] = s.makeResolutionWithPrevious(prev, fA, fB, relName)
		}
	}

	return mth, dif
}

func (s *synchronize) makeResolutionWithPrevious(prev, fA, fB dirreader.FileInfo, relName string) *resolution {
	switch {
	case fA.Name != "" && fB.Name == "":
		if fA.ModTime.Equal(prev.ModTime) {
			return makeResolutionRemove(fA, s.pathA, relName)
		}
		fB.PathAbs = filepath.Join(s.pathB, relName)
	case fA.Name == "" && fB.Name != "":
		if fB.ModTime.Equal(prev.ModTime) {
			return makeResolutionRemove(fB, s.pathB, relName)
		}
		fA.PathAbs = filepath.Join(s.pathA, relName)
	default:
		if fA.ModTime.After(prev.ModTime) && fB.ModTime.Equal(prev.ModTime) {
			return makeResolutionReplace(fA, fB)
		}
		if fB.ModTime.After(prev.ModTime) && fA.ModTime.Equal(prev.ModTime) {
			return makeResolutionReplace(fB, fA)
		}
	}
	return makeResolutionQuestion(fA, fB)
}
