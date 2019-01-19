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
	String() string
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
		_ = os.Remove(b.FiDst.Abs)
		return fmt.Errorf("could not copy file %v: %v", b.FiDst.Abs, err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("could not close file %v: %v", b.FiDst.Abs, err)
	}
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

func (b *Base) String() string {
	return fmt.Sprintf("%q\t%v\t%q\n\t%v %v\n\t%v %q\n\t%v %q",
		b.FiSrc.Abs, "match", b.FiDst.Abs, "Size", b.FiDst.Size,
		"Time", b.FiDst.Time, "Hash", b.FiDst.Hash)
}

func (c *Create) String() string {
	return fmt.Sprintf("%q\t%v\t%q\t%q\t%q",
		c.FiSrc.Name, c.FiSrc.Size, c.FiSrc.Time, "the file will be creating in", c.FiDst.Abs)
}

func (d *Delete) String() string {
	return fmt.Sprintf("%q\t%v\t%q\t%q\t%q",
		d.FiDst.Name, d.FiDst.Size, d.FiDst.Time, "the file will be deleting", d.FiDst.Abs)
}

func (r *Replace) String() string {
	return fmt.Sprintf("%q\t%v\t%q\t%q\t%q",
		r.FiSrc.Name, r.FiSrc.Size, r.FiSrc.Time, "the file will be replacing", r.FiDst.Abs)
}

type Act int

const (
	ActMatch = Act(iota)
	ActDelete
	ActProblem
	ActReplace
	ActCreate
)

func (a Act) String() string {
	switch a {
	case ActMatch:
		return "Match"
	case ActDelete:
		return "Delete"
	case ActProblem:
		return "Problem"
	case ActReplace:
		return "Replace"
	case ActCreate:
		return "Create"
	}
	return "Unknown action"
}

type Resolution struct {
	Act   Act
	FiSrc FI
	Dst   string
}

func (r *Resolution) String() string {
	var s string
	switch r.Act {
	case ActMatch:
		s = fmt.Sprintf("%v %v", r.FiSrc.Abs, "has not been changed")
	case ActDelete:
		s = fmt.Sprintf("%v has been deleted", r.FiSrc.Abs)
	case ActReplace:
		s = fmt.Sprintf("%v %v %q", r.FiSrc.Abs, "has been changed", r.FiSrc.Time)
	case ActProblem:
		s = fmt.Sprintf("%v has time earlier than the previous synchronization", r.FiSrc.Abs)
	case ActCreate:
		s = fmt.Sprintf("%v %v %q", r.FiSrc.Abs, "was created, time", r.FiSrc.Time)
	default:
		return "Unknown resolution"
	}
	return s
}

func CompareSync(arr1, arr2 []FI, abs string) map[string]Resolution {
	m1 := toMap(arr1)
	m2 := toMap(arr2)
	m := make(map[string]Resolution)
	for rel, fi1 := range m1 {
		if fi2, ok := m2[rel]; ok && fi1.Time.Equal(fi2.Time) {
			m[rel] = Resolution{
				Act:   ActMatch,
				FiSrc: fi2,
				Dst:   filepath.Join(abs, rel),
			}
		}
		if fi2, ok := m2[rel]; !ok {
			m[rel] = Resolution{
				Act:   ActDelete,
				FiSrc: fi1,
				Dst:   filepath.Join(abs, rel),
			}
		} else if !fi1.Dir && !fi2.Dir {
			switch {
			case fi1.Time.After(fi2.Time):
				m[rel] = Resolution{
					Act:   ActProblem,
					FiSrc: fi2,
					Dst:   filepath.Join(abs, rel),
				}
			case fi2.Time.After(fi1.Time):
				m[rel] = Resolution{
					Act:   ActReplace,
					FiSrc: fi2,
					Dst:   filepath.Join(abs, rel),
				}
			}
		}
	}
	for rel, fi2 := range m2 {
		if _, ok := m1[rel]; !ok {
			m[rel] = Resolution{
				Act:   ActCreate,
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
			switch res1.Act {
			case ActMatch:
				switch res2.Act {
				case ActMatch:
					match = append(match, createAction(res1.Act, res1.FiSrc, res2.FiSrc))
				case ActDelete:
					dfr = append(dfr, createAction(res2.Act, res2.FiSrc, res1.FiSrc))
				case ActReplace:
					dfr = append(dfr, createAction(res2.Act, res2.FiSrc, res1.FiSrc))
				default:
					dfr = append(dfr, question(res1, res2))
				}
			case ActDelete:
				switch res2.Act {
				case ActMatch:
					dfr = append(dfr, createAction(res1.Act, res1.FiSrc, res2.FiSrc))
				case ActDelete:
					continue
				default:
					dfr = append(dfr, question(res1, res2))
				}
			case ActProblem:
				if res2.Act == ActProblem && res1.FiSrc.Time == res2.FiSrc.Time {
					match = append(match, createAction(ActMatch, res1.FiSrc, res2.FiSrc))
				} else {
					dfr = append(dfr, question(res1, res2))
				}
			case ActReplace:
				switch res2.Act {
				case ActMatch:
					dfr = append(dfr, createAction(res1.Act, res1.FiSrc, res2.FiSrc))
				case ActReplace:
					if res1.FiSrc.Time == res2.FiSrc.Time {
						match = append(match, createAction(ActMatch, res1.FiSrc, res2.FiSrc))
					} else {
						dfr = append(dfr, question(res1, res2))
					}
				default:
					dfr = append(dfr, question(res1, res2))
				}
			case ActCreate:
				if res1.FiSrc.Time == res2.FiSrc.Time {
					match = append(match, createAction(ActMatch, res1.FiSrc, res2.FiSrc))
				} else {
					dfr = append(dfr, question(res1, res2))
				}
			}
		} else {
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
	fmt.Println("File", res1.String(), "and file", res2.String())
	if res1.Act == ActMatch {
		res1.Act = ActReplace
	}
	if res1.Act == ActProblem {
		if res2.Act == ActDelete {
			res1.Act = ActCreate
		} else {
			res1.Act = ActReplace
		}
	}
	if res2.Act == ActMatch {
		res2.Act = ActReplace
	}
	if res2.Act == ActProblem {
		if res1.Act == ActDelete {
			res2.Act = ActCreate
		} else {
			res2.Act = ActReplace
		}
	}
	if res1.Act == ActCreate && res2.Act == ActCreate {
		res1.Act = ActReplace
		res2.Act = ActReplace
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

//TODO: convert to method
func createAction(act Act, fiSrc, fiDst FI) Action {
	var action Action
	switch act {
	case ActMatch:
		action = &Base{
			FiSrc: fiSrc,
			FiDst: fiDst,
		}
	case ActDelete:
		action = &Delete{Base{
			FiSrc: fiSrc,
			FiDst: fiDst,
		}}
	case ActReplace:
		action = &Replace{Base{
			FiSrc: fiSrc,
			FiDst: fiDst,
		}}
	case ActCreate:
		action = &Create{Base{
			FiSrc: fiSrc,
			FiDst: fiDst,
		}}
	}
	return action
}

func toMap(arr []FI) map[string]FI {
	m := make(map[string]FI)
	for _, fi := range arr {
		//TODO: некоторые функции вроде os.Stat могут вернуть полный путь из info.Name()
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
