package synchronize

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gromey/filemanager/dirreader"
)

type base struct {
	fiSrc dirreader.FileInfo
	fiDst dirreader.FileInfo
}

func (b *base) Apply() error {
	f, err := os.Open(b.fiSrc.PathAbs)
	if err != nil {
		return fmt.Errorf("could not open %s: %s", b.fiSrc.PathAbs, err)
	}
	defer f.Close()

	var fi os.FileInfo
	if fi, err = f.Stat(); err == nil && fi.IsDir() {
		return nil
	}

	path := filepath.Dir(b.fiDst.PathAbs)
	if err = os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("could not create Dir %s: %s", path, err)
	}

	if err = b.createAndCopy(f); err != nil {
		return err
	}

	if err = os.Chtimes(b.fiDst.PathAbs, b.fiSrc.ModTime(), b.fiSrc.ModTime()); err != nil {
		return fmt.Errorf("could not change ModTime %s: %s", b.fiDst.PathAbs, err)
	}

	return nil
}

func (b *base) createAndCopy(f *os.File) error {
	w, err := os.Create(b.fiDst.PathAbs)
	if err != nil {
		return fmt.Errorf("could not create file %s: %s", b.fiDst.PathAbs, err)
	}

	defer func() {
		if err = w.Close(); err != nil {
			err = fmt.Errorf("could not close file %s: %s", b.fiDst.PathAbs, err)
		}
	}()

	if _, err = io.Copy(w, f); err != nil {
		_ = os.Remove(b.fiDst.PathAbs)
		return fmt.Errorf("could not copy file %s: %s", b.fiDst.PathAbs, err)
	}

	return err
}

//func (b *base) String() string {
//	return fmt.Sprintf("%q\t%v\t%q\n\t%v %v\n\t%v %q\n\t%v %q",
//		b.fiSrc.PathAbs, "match", b.fiDst.PathAbs, "Size", b.fiDst.Size,
//		"ModTime", b.fiDst.ModTime, "Hash", b.fiDst.Hash)
//}

type Create struct {
	base
}

//func (c *Create) String() string {
//	return fmt.Sprintf("%q\t%v\t%q\t%q\t%q",
//		c.fiSrc.Name, c.fiSrc.Size, c.fiSrc.ModTime, "the file will be creating in", c.fiDst.PathAbs)
//}

type Replace struct {
	base
}

//func (r *Replace) String() string {
//	return fmt.Sprintf("%q\t%v\t%q\t%q\t%q",
//		r.fiSrc.Name, r.fiSrc.Size, r.fiSrc.ModTime, "the file will be replacing", r.fiDst.PathAbs)
//}

type Delete struct {
	base
}

func (d *Delete) Apply() error {
	if err := os.Remove(d.fiDst.PathAbs); err != nil {
		return fmt.Errorf("could not delete file %s: %s", d.fiDst.PathAbs, err)
	}
	return nil
}

//func (d *Delete) String() string {
//	return fmt.Sprintf("%q\t%v\t%q\tthe file will be deleting\t%s",
//		d.fiDst.Name(), d.fiDst.Size(), d.fiDst.ModTime(), d.fiDst.PathAbs)
//}

func compareFileInfo(res, included []dirreader.FileInfo, abs string) map[string]*resolution {
	m1, m2 := toMap(res), toMap(included)

	m := make(map[string]*resolution)

	for relName, fi1 := range m1 {
		if fi2, ok := m2[relName]; ok && !fi1.IsDir() && !fi2.IsDir() {
			switch {
			case fi1.ModTime().Equal(fi2.ModTime()):
				m[relName] = makeResolutionMatch(fi1, fi2)
			case fi1.ModTime().Before(fi2.ModTime()):
				m[relName] = makeResolutionReplace(fi2, abs, relName)
			case fi1.ModTime().After(fi2.ModTime()):
				m[relName] = makeResolutionProblem(fi2, abs, relName)
			}
		}

		if _, ok := m2[relName]; !ok && !fi1.IsDir() {
			m[relName] = makeResolutionDelete(fi1, abs, relName)
		}
	}

	for relName, fi2 := range m2 {
		if _, ok := m1[relName]; !ok {
			m[relName] = makeResolutionCreate(fi2, abs, relName)
		}
	}

	return m
}

func toMap(arr []dirreader.FileInfo) map[string]dirreader.FileInfo {
	m := make(map[string]dirreader.FileInfo, len(arr))
	for _, fi := range arr {
		m[filepath.Join(fi.PathRel, fi.Name())] = fi
	}
	return m
}

func CompareResolution(res1, res2 map[string]resolution) {

}

func compareResolution(m1, m2 map[string]resolution) ([]Action, []Action) {
	var match, dfr []Action
	for relName, res1 := range m1 {
		if res2, ok := m2[relName]; ok {
			switch res1.action {
			case actMatch:
				switch res2.action {
				case actMatch:
					match = append(match, res2.createAction())
				case actDelete, actReplace:
					dfr = append(dfr, res2.createAction())
				default:
					dfr = append(dfr, question(res1, res2))
				}
			case actDelete:
				switch res2.action {
				case actMatch:
					dfr = append(dfr, res1.createAction())
				case actDelete:
					continue
				default:
					dfr = append(dfr, question(res1, res2))
				}
			case actProblem:
				if res2.action == actProblem && res1.fiSrc.ModTime() == res2.fiSrc.ModTime() {
					res := resolution{
						action: actMatch, base: base{
							fiSrc: res1.fiSrc,
							fiDst: res2.fiSrc,
						}}
					match = append(match, res.createAction())
				} else {
					dfr = append(dfr, question(res1, res2))
				}
			case actReplace:
				switch res2.action {
				case actMatch:
					dfr = append(dfr, res1.createAction())
				case actReplace:
					if res1.fiSrc.ModTime() == res2.fiSrc.ModTime() {
						res := resolution{
							action: actMatch, base: base{
								fiSrc: res1.fiSrc,
								fiDst: res2.fiSrc,
							}}
						match = append(match, res.createAction())
					} else {
						dfr = append(dfr, question(res1, res2))
					}
				default:
					dfr = append(dfr, question(res1, res2))
				}
			case actCreate:
				if res1.fiSrc.ModTime() == res2.fiSrc.ModTime() {
					res := resolution{
						action: actMatch, base: base{
							fiSrc: res1.fiSrc,
							fiDst: res2.fiSrc,
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

func question(res1, res2 resolution) Action {
	fmt.Println("File", res1.String(), "and file", res2.String())
	if res1.action == actMatch {
		res1.action = actReplace
	}
	if res1.action == actProblem {
		if res2.action == actDelete {
			res1.action = actCreate
		} else {
			res1.action = actReplace
		}
	}
	if res2.action == actMatch {
		res2.action = actReplace
	}
	if res2.action == actProblem {
		if res1.action == actDelete {
			res2.action = actCreate
		} else {
			res2.action = actReplace
		}
	}
	if res1.action == actCreate && res2.action == actCreate {
		res1.action = actReplace
		res2.action = actReplace
	}
	fmt.Println("Enter \">\" for", res1.action.String(), "file in", res1.fiDst.PathAbs+
		" or enter \"<\" for", res2.action.String(), "file in", res2.fiDst.PathAbs+
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
