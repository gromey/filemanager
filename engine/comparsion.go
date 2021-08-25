package engine

import (
	"fmt"
	"github.com/gromey/filemanager/pkg/dirreader"
	"io"
	"os"
	"path/filepath"
)

type Base struct {
	FiSrc dirreader.FileInfo
	FiDst dirreader.FileInfo
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
	r, err := os.Open(b.FiSrc.PathAbs)
	if err != nil {
		return fmt.Errorf("could not open %v: %v", b.FiSrc.PathAbs, err)
	}
	defer r.Close()
	if fi, err := r.Stat(); err == nil && fi.IsDir() {
		return nil
	}
	path := filepath.Dir(b.FiDst.PathAbs)
	if err = os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("could not create Dir %v: %v", path, err)
	}
	w, err := os.Create(b.FiDst.PathAbs)
	if err != nil {
		return fmt.Errorf("could not create file %v: %v", b.FiDst.PathAbs, err)
	}
	defer w.Close()
	if _, err = io.Copy(w, r); err != nil {
		_ = os.Remove(b.FiDst.PathAbs)
		return fmt.Errorf("could not copy file %v: %v", b.FiDst.PathAbs, err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("could not close file %v: %v", b.FiDst.PathAbs, err)
	}
	if err = os.Chtimes(b.FiDst.PathAbs, b.FiSrc.ModTime(), b.FiSrc.ModTime()); err != nil {
		return fmt.Errorf("could not change ModTime %v: %v", b.FiDst.PathAbs, err)
	}
	return nil
}

func (d *Delete) Apply() error {
	err := os.Remove(d.FiDst.PathAbs)
	if err != nil {
		return fmt.Errorf("could not delete file %v: %v", d.FiDst.PathAbs, err)
	}
	return nil
}

func (b *Base) String() string {
	return fmt.Sprintf("%q\t%v\t%q\n\t%v %v\n\t%v %q\n\t%v %q",
		b.FiSrc.PathAbs, "match", b.FiDst.PathAbs, "Size", b.FiDst.Size,
		"ModTime", b.FiDst.ModTime, "Hash", b.FiDst.Hash)
}

func (c *Create) String() string {
	return fmt.Sprintf("%q\t%v\t%q\t%q\t%q",
		c.FiSrc.Name, c.FiSrc.Size, c.FiSrc.ModTime, "the file will be creating in", c.FiDst.PathAbs)
}

func (d *Delete) String() string {
	return fmt.Sprintf("%q\t%v\t%q\t%q\t%q",
		d.FiDst.Name, d.FiDst.Size, d.FiDst.ModTime, "the file will be deleting", d.FiDst.PathAbs)
}

func (r *Replace) String() string {
	return fmt.Sprintf("%q\t%v\t%q\t%q\t%q",
		r.FiSrc.Name, r.FiSrc.Size, r.FiSrc.ModTime, "the file will be replacing", r.FiDst.PathAbs)
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
	Act Act
	Base
}

func (r *Resolution) String() string {
	var s string
	switch r.Act {
	case ActMatch:
		s = fmt.Sprintf("%v %v", r.FiSrc.PathAbs, "has not been changed")
	case ActDelete:
		s = fmt.Sprintf("%v has been deleted", r.FiSrc.PathAbs)
	case ActReplace:
		s = fmt.Sprintf("%v %v %q", r.FiSrc.PathAbs, "has been changed", r.FiSrc.ModTime)
	case ActProblem:
		s = fmt.Sprintf("%v has time earlier than the previous synchronization", r.FiSrc.PathAbs)
	case ActCreate:
		s = fmt.Sprintf("%v %v %q", r.FiSrc.PathAbs, "was created, time", r.FiSrc.ModTime)
	default:
		return "Unknown resolution"
	}
	return s
}

func CompareSync(arr1, arr2 []dirreader.FileInfo, abs string) map[string]Resolution {
	m1 := toMap(arr1)
	m2 := toMap(arr2)
	m := make(map[string]Resolution)
	for relName, fi1 := range m1 {
		if fi2, ok := m2[relName]; ok && fi1.ModTime().Equal(fi2.ModTime()) {
			m[relName] = Resolution{
				Act: ActMatch,
				Base: Base{
					FiSrc: fi1,
					FiDst: fi2,
				},
			}
		}
		if fi2, ok := m2[relName]; !ok {
			m[relName] = Resolution{
				Act: ActDelete,
				Base: Base{
					FiSrc: fi1,
					FiDst: dirreader.FileInfo{
						PathAbs: filepath.Join(abs, relName),
					},
				},
			}
		} else if !fi1.IsDir() && !fi2.IsDir() {
			base := Base{
				FiSrc: fi2,
				FiDst: dirreader.FileInfo{
					PathAbs: filepath.Join(abs, relName),
				},
			}
			switch {
			case fi1.ModTime().After(fi2.ModTime()):
				m[relName] = Resolution{
					Act:  ActProblem,
					Base: base,
				}
			case fi2.ModTime().After(fi1.ModTime()):
				m[relName] = Resolution{
					Act:  ActReplace,
					Base: base,
				}
			}
		}
	}
	for relName, fi2 := range m2 {
		if _, ok := m1[relName]; !ok {
			m[relName] = Resolution{
				Act: ActCreate,
				Base: Base{
					FiSrc: fi2,
					FiDst: dirreader.FileInfo{
						PathAbs: filepath.Join(abs, relName),
					},
				},
			}
		}
	}
	return m
}

func CompareResolution(m1, m2 map[string]Resolution) ([]Action, []Action) {
	var match, dfr []Action
	for relName, res1 := range m1 {
		if res2, ok := m2[relName]; ok {
			switch res1.Act {
			case ActMatch:
				switch res2.Act {
				case ActMatch:
					match = append(match, res2.createAction())
				case ActDelete, ActReplace:
					dfr = append(dfr, res2.createAction())
				default:
					dfr = append(dfr, question(res1, res2))
				}
			case ActDelete:
				switch res2.Act {
				case ActMatch:
					dfr = append(dfr, res1.createAction())
				case ActDelete:
					continue
				default:
					dfr = append(dfr, question(res1, res2))
				}
			case ActProblem:
				if res2.Act == ActProblem && res1.FiSrc.ModTime() == res2.FiSrc.ModTime() {
					res := Resolution{
						Act: ActMatch, Base: Base{
							FiSrc: res1.FiSrc,
							FiDst: res2.FiSrc,
						}}
					match = append(match, res.createAction())
				} else {
					dfr = append(dfr, question(res1, res2))
				}
			case ActReplace:
				switch res2.Act {
				case ActMatch:
					dfr = append(dfr, res1.createAction())
				case ActReplace:
					if res1.FiSrc.ModTime() == res2.FiSrc.ModTime() {
						res := Resolution{
							Act: ActMatch, Base: Base{
								FiSrc: res1.FiSrc,
								FiDst: res2.FiSrc,
							}}
						match = append(match, res.createAction())
					} else {
						dfr = append(dfr, question(res1, res2))
					}
				default:
					dfr = append(dfr, question(res1, res2))
				}
			case ActCreate:
				if res1.FiSrc.ModTime() == res2.FiSrc.ModTime() {
					res := Resolution{
						Act: ActMatch, Base: Base{
							FiSrc: res1.FiSrc,
							FiDst: res2.FiSrc,
						}}
					match = append(match, res.createAction())
				} else {
					dfr = append(dfr, question(res1, res2))
				}
			}
		} else {
			dfr = append(dfr, res1.createAction())
		}
	}
	for relName, res2 := range m2 {
		if _, ok := m1[relName]; !ok {
			dfr = append(dfr, res2.createAction())
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
	fmt.Println("Enter \">\" for", res1.Act.String(), "file in", res1.FiDst.PathAbs+
		" or enter \"<\" for", res2.Act.String(), "file in", res2.FiDst.PathAbs+
		" or enter any other character to skip the file synchronization\n")
	var ask string
	fmt.Scanln(&ask)
	if ask == "<" {
		return res2.createAction()
	}
	if ask == ">" {
		return res1.createAction()
	} else {
		fmt.Println("canceled")
		return nil
	}
}

func (r *Resolution) createAction() Action {
	var action Action
	switch r.Act {
	case ActMatch:
		action = &r.Base
	case ActDelete:
		action = &Delete{r.Base}
	case ActReplace:
		action = &Replace{r.Base}
	case ActCreate:
		action = &Create{r.Base}
	}
	return action
}

func toMap(arr []dirreader.FileInfo) map[string]dirreader.FileInfo {
	m := make(map[string]dirreader.FileInfo)
	for _, fi := range arr {
		//TODO: некоторые функции вроде os.Stat могут вернуть полный путь из info.Name()
		m[filepath.Join(fi.PathRel, fi.Name())] = fi
	}
	return m
}

func CompareDpl(arr []dirreader.FileInfo) []Action {
	m := make(map[string]dirreader.FileInfo)
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
