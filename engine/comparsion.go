package engine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Base struct {
	Fi  FI
	Dst string
}

type Create struct {
	Base
}

type Replace struct {
	Base
}

type Action interface {
	Apply() error
	Description() string
}

func (c *Base) Apply() error {
	s, err := os.Open(c.Fi.Abs)
	if err != nil {
		return fmt.Errorf("could not open %v: %v", c.Fi.Abs, err)
	}
	defer s.Close()
	if fi, err := s.Stat(); err == nil && fi.IsDir() {
		return nil
	}
	path := filepath.Dir(c.Dst)
	if err = os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("could not create Dir %v: %v", path, err)
	}
	w, err := os.Create(c.Dst)
	if err != nil {
		return fmt.Errorf("could not create file %v: %v", c.Dst, err)
	}
	defer w.Close()
	if _, err = io.Copy(w, s); err != nil {
		return fmt.Errorf("could not copy file %v: %v", c.Dst, err)
	}
	w.Close()
	if err = os.Chtimes(c.Dst, c.Fi.Time, c.Fi.Time); err != nil {
		return fmt.Errorf("could not change Time %v: %v", c.Dst, err)
	}
	return nil
}

func (c *Base) Description() string {
	return fmt.Sprintf("%q\t%v\t%q\t%v\t%q",
		c.Fi.Abs, "match", c.Dst, c.Fi.Size, c.Fi.Hash)
}

func (c *Create) Description() string {
	return fmt.Sprintf("%q\t%v\t%q\t%q\t%q",
		c.Fi.Name, c.Fi.Size, c.Fi.Time, "the file need to create in", c.Dst)
}

func (r *Replace) Description() string {
	return fmt.Sprintf("%q\t%v\t%q\t%q\t%q",
		r.Fi.Name, r.Fi.Size, r.Fi.Time, "the file need to repleace", r.Dst)

}

func CompareSync(arr1, arr2 []FI, abs1, abs2 string) ([]Action, []Action) {
	m1 := Convert(arr1)
	m2 := Convert(arr2)
	var match, dfr []Action
	for rel, fi1 := range m1 {
		if fi2, ok := m2[rel]; ok && fi1.Time.Equal(fi2.Time) {
			match = append(match, &Base{
				Fi:  fi1,
				Dst: fi2.Abs,
			})
		}
		if fi2, ok := m2[rel]; !ok {
			dfr = append(dfr, &Create{Base{
				Fi:  fi1,
				Dst: filepath.Join(abs2, rel),
			}})
		} else if !fi1.Dir && !fi2.Dir {
			switch {
			case fi1.Time.After(fi2.Time):
				dfr = append(dfr, &Replace{Base{
					Fi:  fi1,
					Dst: fi2.Abs,
				}})
			case fi2.Time.After(fi1.Time):
				dfr = append(dfr, &Replace{Base{
					Fi:  fi2,
					Dst: fi1.Abs,
				}})
			}
		}
	}
	for rel, fi2 := range m2 {
		if _, ok := m1[rel]; !ok {
			dfr = append(dfr, &Create{Base{
				Fi:  fi2,
				Dst: filepath.Join(abs1, rel),
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

func CompareDupl(arr []FI) []Action {
	m := make(map[string]FI)
	var match []Action
	for _, fi := range arr {
		if fiM, ok := m[fi.Hash]; ok {
			match = append(match, &Base{
				Fi:  fi,
				Dst: fiM.Abs,
			})
		} else {
			m[fi.Hash] = fi
		}
	}
	return match
}
