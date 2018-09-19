package mode

import (
	"fmt"
	"sort"
	"syncdata/engine"
)

func Dupl(c Config) error {
	var ext []string
	if c.Mask.On {
		ext = c.Mask.Ext
	}
	var excl, incl []engine.FI
	for _, path := range c.Paths {
		ex, in, err := engine.ReadDir(path, ext)
		if err != nil {
			return err
		}
		excl = append(excl, ex...)
		incl = append(incl, in...)
	}
	if c.Mask.On && c.Mask.Include {
		excl, incl = incl, excl
	}
	if c.Mask.On && c.Mask.Verbose {
		sort.Slice(excl, func(i, j int) bool {
			return excl[i].Abs < excl[j].Abs
		})
		for _, fi := range excl {
			fmt.Printf("%q\t%v\t%q\t%q\n", fi.Abs, fi.Size, fi.Time, "the file is excluded by a mask")
		}
	}
	match := engine.CompareDpl(incl)
	if len(match) == 0 {
		fmt.Println("No match\n")
		return nil
	}
	for _, action := range match {
		fmt.Println(action.Description())
	}
	return nil
}
