package main

import (
	"fmt"
	"github.com/gromey/filemanager/pkg/dirreader"
)

func main() {
	dr := dirreader.SetDirReader("/home/evgeniy/Finndon/", []string{"json", "csv", "pem"}, true)
	incl, excl, err := dr.Exec()
	if err != nil {
		fmt.Println(err)
	}
	for _, fi := range incl {
		fmt.Printf("%#v\n", fi)
	}
	fmt.Println()
	for _, fi := range excl {
		fmt.Printf("%#v\n", fi)
	}
}
