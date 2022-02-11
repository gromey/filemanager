package duplicate

import (
	"github.com/gromey/filemanager/dirreader"
	"sort"
	"strings"
	"time"
)

type Test struct {
	Hash    string
	Size    int64
	ModTime time.Time
	Paths   []string
}

func compare(arr []dirreader.FileInfo) []Test {
	m := make(map[string]dirreader.FileInfo, len(arr))
	match := make(map[string]Test)

	for _, fi := range arr {
		if fiDuplicate, ok := m[fi.Hash]; !ok {
			m[fi.Hash] = fi
		} else {
			if dup, ok := match[fi.Hash]; !ok {
				match[fi.Hash] = Test{
					Hash:    fi.Hash,
					Size:    fi.Size(),
					ModTime: fi.ModTime(),
					Paths:   []string{strings.TrimPrefix(fi.PathAbs, "/"), strings.TrimPrefix(fiDuplicate.PathAbs, "/")},
				}
			} else {
				dup.Paths = append(dup.Paths, strings.TrimPrefix(fi.PathAbs, "/"))
				match[fi.Hash] = dup
			}
		}
	}

	res := make([]Test, len(match))
	i := 0
	for _, dup := range match {
		sort.Slice(dup.Paths, func(i, j int) bool {
			return dup.Paths[i] < dup.Paths[j]
		})
		res[i] = dup
		i++
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Hash < res[j].Hash
	})

	return res
}
