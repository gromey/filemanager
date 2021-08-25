package reject

//
//import (
//	"encoding/json"
//	"fmt"
//	"github.com/GroM1124/filemanager/pkg/readdir"
//	"io/ioutil"
//	"os"
//	"sort"
//	"strings"
//)
//
//type Reject struct {
//	Paths []string `json:"paths"`
//	Mask  struct {
//		On      bool     `json:"on"`
//		Ext     []string `json:"ext"`
//		Include bool     `json:"Include"`
//		Verbose bool     `json:"verbose"`
//	} `json:"mask"`
//	Delete []string `json:"delete"`
//	Space  []string `json:"space"`
//}
//
//func Run(config string) error {
//	data, err := ioutil.ReadFile(config)
//	if os.IsNotExist(err) {
//		return fmt.Errorf("no config file")
//	} else if err != nil {
//		return fmt.Errorf("could not read config %v", err)
//	}
//	var d []Reject
//	err = json.Unmarshal(data, &d)
//	if err != nil {
//		return fmt.Errorf("could not unmarshal config %v", err)
//	}
//	for _, reject := range d {
//		err := reject.Scanner()
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//func (r *Reject) Scanner() error {
//	var ext []string
//	if r.Mask.On {
//		ext = r.Mask.Ext
//	}
//	var excl, incl []readdir.FileInfo
//	for _, path := range r.Paths {
//		rd := readdir.SetReadDir(path, ext, true)
//		ex, in, err := rd.ReadDirectory()
//		if err != nil {
//			return err
//		}
//		excl = append(excl, ex...)
//		incl = append(incl, in...)
//	}
//	if r.Mask.On && r.Mask.Include {
//		excl, incl = incl, excl
//	}
//	if r.Mask.On && r.Mask.Verbose {
//		sort.Slice(excl, func(i, j int) bool {
//			return excl[i].PathAbs < excl[j].PathAbs
//		})
//		for _, fi := range excl {
//			fmt.Printf("%q\t%v\t%q\t%q\n", fi.PathAbs, fi.Size, fi.ModTime, "the file is excluded by a mask")
//		}
//	}
//
//	for _, fi := range incl {
//		if err := r.NameEditor(fi); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (r *Reject) NameEditor(fi readdir.FileInfo) error {
//	path := strings.Trim(fi.PathAbs, fi.Name)
//	name := fi.Name
//	for _, str := range r.Delete {
//		name = strings.Replace(name, str, "", 1)
//	}
//	for _, str := range r.Space {
//		name = strings.Replace(name, str, " ", len(name))
//	}
//	dst := strings.Join([]string{path, name}, "")
//	fmt.Println(dst)
//	if err := os.Rename(fi.PathAbs, dst); err != nil {
//		return fmt.Errorf("could not rename file %v: %v", fi.PathAbs, err)
//	}
//	return nil
//}
