package engine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Base struct {
	FiSrc FI
	FiDst FI
}

type Create struct {
	Base
}

type Delete struct {
	Base
}

type Replace struct {
	Base
}

type Action interface {
	Apply() error
	Description() string
}

func (b *Base) Apply() error {
	s, err := os.Open(b.FiSrc.Abs)
	if err != nil {
		return fmt.Errorf("could not open %v: %v", b.FiSrc.Abs, err)
	}
	defer s.Close()
	if fi, err := s.Stat(); err == nil && fi.IsDir() {
		return nil
	}
	path := filepath.Dir(b.FiDst.Abs)
	if err = os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("could not create Dir %v: %v", path, err)
	}
	w, err := os.Create(b.FiDst.Abs)
	if err != nil {
		return fmt.Errorf("could not create file %v: %v", b.FiDst.Abs, err)
	}
	defer w.Close()
	if _, err = io.Copy(w, s); err != nil {
		return fmt.Errorf("could not copy file %v: %v", b.FiDst.Abs, err)
	}
	w.Close()
	if err = os.Chtimes(b.FiDst.Abs, b.FiSrc.Time, b.FiSrc.Time); err != nil {
		return fmt.Errorf("could not change Time %v: %v", b.FiDst.Abs, err)
	}
	return nil
}

func (d *Delete) Apply() error {
	err := os.Remove(d.FiDst.Abs)
	if err != nil {
		return fmt.Errorf("could not delete file %v: %v", d.FiDst.Abs, err)
	}
	return nil
}

func (b *Base) Description() string {
	return fmt.Sprintf("%q\t%v\t%q\n\t%v = %v\n\t%q = %q\n\t%q = %q",
		b.FiSrc.Abs, "match", b.FiDst.Abs, b.FiSrc.Size, b.FiDst.Size,
		b.FiSrc.Time, b.FiDst.Time, b.FiSrc.Hash, b.FiDst.Hash)
}

func (c *Create) Description() string {
	return fmt.Sprintf("%q\t%v\t%q\t%q\t%q",
		c.FiSrc.Name, c.FiSrc.Size, c.FiSrc.Time, "the file will be creating in", c.FiDst.Abs)
}

func (d *Delete) Description() string {
	return fmt.Sprintf("%q\t%v\t%q\t%q\t%q",
		d.FiDst.Name, d.FiDst.Size, d.FiDst.Time, "the file will be deleting", d.FiDst.Abs)
}

func (r *Replace) Description() string {
	return fmt.Sprintf("%q\t%v\t%q\t%q\t%q",
		r.FiSrc.Name, r.FiSrc.Size, r.FiSrc.Time, "the file will be replacing", r.FiDst.Abs)

}

func CompareSync(arr1, arr2 []FI, abs1, abs2 string) ([]Action, []Action) {
	m1 := Convert(arr1)
	m2 := Convert(arr2)
	var match, dfr []Action
	for rel, fi1 := range m1 {
		if fi2, ok := m2[rel]; ok && fi1.Time.Equal(fi2.Time) {
			match = append(match, &Base{
				FiSrc: fi1,
				FiDst: fi2,
			})
		}
		if fi2, ok := m2[rel]; !ok {
			dfr = append(dfr, &Create{Base{
				FiSrc: fi1,
				FiDst: FI{
					Abs: filepath.Join(abs2, rel),
				},
			}})
		} else if !fi1.Dir && !fi2.Dir {
			switch {
			case fi1.Time.After(fi2.Time):
				dfr = append(dfr, &Replace{Base{
					FiSrc: fi1,
					FiDst: fi2,
				}})
			case fi2.Time.After(fi1.Time):
				dfr = append(dfr, &Replace{Base{
					FiSrc: fi2,
					FiDst: fi1,
				}})
			}
		}
	}
	for rel, fi2 := range m2 {
		if _, ok := m1[rel]; !ok {
			dfr = append(dfr, &Create{Base{
				FiSrc: fi2,
				FiDst: FI{
					Abs: filepath.Join(abs1, rel),
				},
			}})
		}
	}
	return match, dfr
}

func Convert(arr []FI) map[string]FI {
	m := make(map[string]FI)
	for _, fi := range arr {
		m[filepath.Join(fi.Rel, fi.Name)] = fi
	}
	return m
}

func CompareDpl(arr []FI) []Action {
	m := make(map[string]FI)
	var match []Action
	for _, fi := range arr {
		if fiM, ok := m[fi.Hash]; ok {
			match = append(match, &Base{
				FiSrc: fi,
				FiDst: fiM,
			})
		} else {
			m[fi.Hash] = fi
		}
	}
	return match
}
