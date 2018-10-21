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

type Crt struct {
	Base
}

type Dlt struct {
	Base
}

type Rpl struct {
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

func (d *Dlt) Apply() error {
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

func (c *Crt) Description() string {
	return fmt.Sprintf("%q\t%v\t%q\t%q\t%q",
		c.FiSrc.Name, c.FiSrc.Size, c.FiSrc.Time, "the file will be creating in", c.FiDst.Abs)
}

func (d *Dlt) Description() string {
	return fmt.Sprintf("%q\t%v\t%q\t%q\t%q",
		d.FiDst.Name, d.FiDst.Size, d.FiDst.Time, "the file will be deleting", d.FiDst.Abs)
}

func (r *Rpl) Description() string {
	return fmt.Sprintf("%q\t%v\t%q\t%q\t%q",
		r.FiSrc.Name, r.FiSrc.Size, r.FiSrc.Time, "the file will be replacing", r.FiDst.Abs)
}

type Act int

const (
	Match = Act(iota)
	Delete
	Problem
	Replace
	Create
)

func (a Act) String() string {
	switch a {
	case Match:
		return "Match"
	case Delete:
		return "Delete"
	case Problem:
		return "Problem"
	case Replace:
		return "Replace"
	case Create:
		return "Create"
	}
	return ""
}

type Resolution struct {
	Act   Act
	FiSrc FI
	Dst   string
}

func (r *Resolution) Description() string {
	var d string
	switch r.Act {
	case Match:
		d = fmt.Sprintf("%v %v", r.FiSrc.Abs, "has not been changed")
	case Delete:
		d = fmt.Sprintf("%v has been deleted", r.FiSrc.Abs)
	case Replace:
		d = fmt.Sprintf("%v %v %q", r.FiSrc.Abs, "has been changed", r.FiSrc.Time)
	case Problem:
		d = fmt.Sprintf("%v has time earlier than the previous synchronization", r.FiSrc.Abs)
	case Create:
		d = fmt.Sprintf("%v %v %q", r.FiSrc.Abs, "was created, time", r.FiSrc.Time)
	}
	return d
}

func CompareSync(arr1, arr2 []FI, abs string) map[string]Resolution {
	m1 := Convert(arr1)
	m2 := Convert(arr2)
	m := make(map[string]Resolution)
	for rel, fi1 := range m1 {
		if fi2, ok := m2[rel]; ok && fi1.Time.Equal(fi2.Time) {
			m[rel] = Resolution{
				Act:   Match,
				FiSrc: fi2,
				Dst:   filepath.Join(abs, rel),
			}
		}
		if fi2, ok := m2[rel]; !ok {
			m[rel] = Resolution{
				Act:   Delete,
				FiSrc: fi1,
				Dst:   filepath.Join(abs, rel),
			}
		} else if !fi1.Dir && !fi2.Dir {
			switch {
			case fi1.Time.After(fi2.Time):
				m[rel] = Resolution{
					Act:   Problem,
					FiSrc: fi2,
					Dst:   filepath.Join(abs, rel),
				}
			case fi2.Time.After(fi1.Time):
				m[rel] = Resolution{
					Act:   Replace,
					FiSrc: fi2,
					Dst:   filepath.Join(abs, rel),
				}
			}
		}
	}
	for rel, fi2 := range m2 {
		if _, ok := m1[rel]; !ok {
			m[rel] = Resolution{
				Act:   Create,
				FiSrc: fi2,
				Dst:   filepath.Join(abs, rel),
			}
		}
	}
	return m
}

func CompareResolution(m1, m2 map[string]Resolution) ([]Action, []Action) {
	var match, dfr []Action
	for rel, res1 := range m1 {
		if res2, ok := m2[rel]; ok {
			switch {
			case res1.Act == Match:
				switch {
				case res2.Act == Match:
					match = append(match, createAction(res1.Act, res1.FiSrc, res2.FiSrc))
				case res2.Act == Delete:
					dfr = append(dfr, createAction(res2.Act, res2.FiSrc, res1.FiSrc))
				case res2.Act == Replace:
					dfr = append(dfr, createAction(res2.Act, res2.FiSrc, res1.FiSrc))
				default:
					dfr = append(dfr, question(res1, res2))
				}
			case res1.Act == Delete:
				switch {
				case res2.Act == Match:
					dfr = append(dfr, createAction(res1.Act, res1.FiSrc, res2.FiSrc))
				case res2.Act == Delete:
					continue
				default:
					dfr = append(dfr, question(res1, res2))
				}
			case res1.Act == Problem:
				if res2.Act == Problem && res1.FiSrc.Time == res2.FiSrc.Time {
					match = append(match, createAction(Match, res1.FiSrc, res2.FiSrc))
				} else {
					dfr = append(dfr, question(res1, res2))
				}
			case res1.Act == Replace:
				switch {
				case res2.Act == Match:
					dfr = append(dfr, createAction(res1.Act, res1.FiSrc, res2.FiSrc))
				case res2.Act == Replace:
					if res1.FiSrc.Time == res2.FiSrc.Time {
						match = append(match, createAction(Match, res1.FiSrc, res2.FiSrc))
					} else {
						dfr = append(dfr, question(res1, res2))
					}
				default:
					dfr = append(dfr, question(res1, res2))
				}
			case res1.Act == Create:
				if res1.FiSrc.Time == res2.FiSrc.Time {
					match = append(match, createAction(Match, res1.FiSrc, res2.FiSrc))
				} else {
					dfr = append(dfr, question(res1, res2))
				}
			}
		}
		if _, ok := m2[rel]; !ok {
			dfr = append(dfr, createAction(res1.Act, res1.FiSrc, FI{
				Abs: res1.Dst,
			}))
		}
	}
	for rel, res2 := range m2 {
		if _, ok := m1[rel]; !ok {
			dfr = append(dfr, createAction(res2.Act, res2.FiSrc, FI{
				Abs: res2.Dst,
			}))
		}
	}
	return match, dfr
}

func question(res1, res2 Resolution) Action {
	fmt.Println("File", res1.Description(), "and file", res2.Description())
	if res1.Act == Match {
		res1.Act = Replace
	}
	if res1.Act == Problem {
		if res2.Act == Delete {
			res1.Act = Create
		} else {
			res1.Act = Replace
		}
	}
	if res2.Act == Match {
		res2.Act = Replace
	}
	if res2.Act == Problem {
		if res1.Act == Delete {
			res2.Act = Create
		} else {
			res2.Act = Replace
		}
	}
	if res1.Act == Create && res1.Act == Create {
		res1.Act = Replace
		res2.Act = Replace
	}
	fmt.Println("Enter \"<\" for", res2.Act.String(), "file in", res2.Dst, "or enter \">\" for", res1.Act.String(), "file in", res1.Dst+
		" or enter any other character to skip the file synchronization\n")
	var ask string
	fmt.Scanln(&ask)
	if ask == "<" {
		return createAction(res2.Act, res2.FiSrc, res1.FiSrc)
	}
	if ask == ">" {
		return createAction(res1.Act, res1.FiSrc, res2.FiSrc)
	} else {
		fmt.Println("canceled\n")
		return nil
	}
	return nil
}

func createAction(act Act, fiSrc, fiDst FI) Action {
	var action Action
	switch act {
	case Match:
		action = &Base{
			FiSrc: fiSrc,
			FiDst: fiDst,
		}
	case Delete:
		action = &Dlt{Base{
			FiSrc: fiSrc,
			FiDst: fiDst,
		}}
	case Replace:
		action = &Rpl{Base{
			FiSrc: fiSrc,
			FiDst: fiDst,
		}}
	case Create:
		action = &Crt{Base{
			FiSrc: fiSrc,
			FiDst: fiDst,
		}}
	}
	return action
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
