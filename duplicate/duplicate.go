package duplicate

import (
	"log"
	"sort"
	"strings"
	"time"

	"github.com/gromey/filemanager/dirreader"
)

type Duplicate interface {
	Start() ([]Repeated, error)
}

type Config struct {
	Paths []string `json:"paths"`
	Mask  struct {
		On        bool     `json:"on"`
		Extension []string `json:"extension"`
		Include   bool     `json:"include"`
		Details   bool     `json:"details"`
	} `json:"mask"`
}

func New(c *Config) Duplicate {
	d := &duplicate{
		paths: c.Paths,
	}

	if c.Mask.On {
		d.extension = c.Mask.Extension
		d.include = c.Mask.Include
		d.details = c.Mask.Details
	}

	return d
}

type Repeated struct {
	Hash    string
	Size    int64
	ModTime time.Time
	Paths   []string
}

type duplicate struct {
	paths     []string
	extension []string
	include   bool
	details   bool
}

func (d *duplicate) Start() ([]Repeated, error) {
	var excluded, included []dirreader.FileInfo

	for _, path := range d.paths {
		exclude, include, err := dirreader.New(path, d.extension, d.include, d.details, true).Exec()
		if err != nil {
			return nil, err
		}

		excluded = append(excluded, exclude...)
		included = append(included, include...)
	}

	if d.details {
		log.Printf("%d files was excluded by mask.\n", len(excluded))
		//for _, v := range excluded {
		//	log.Println(v)
		//}
	}

	match := compare(included)
	if len(match) == 0 {
		return nil, nil
	}

	return match, nil
}

func compare(arr []dirreader.FileInfo) []Repeated {
	m := make(map[string]dirreader.FileInfo, len(arr))
	match := make(map[string]Repeated)

	for _, fi := range arr {
		if fiDuplicate, ok := m[fi.Hash]; !ok {
			m[fi.Hash] = fi
		} else {
			var dup Repeated
			if dup, ok = match[fi.Hash]; !ok {
				match[fi.Hash] = Repeated{
					Hash:    fi.Hash,
					Size:    fi.Size(),
					ModTime: fi.ModTime(),
					Paths: []string{
						strings.TrimPrefix(fi.PathAbs, "/"),
						strings.TrimPrefix(fiDuplicate.PathAbs, "/"),
					},
				}
			} else {
				dup.Paths = append(dup.Paths, strings.TrimPrefix(fi.PathAbs, "/"))
				match[fi.Hash] = dup
			}
		}
	}

	res := make([]Repeated, len(match))
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
