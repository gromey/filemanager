package reject

import (
	"fmt"
	"github.com/gromey/filemanager/dirreader"
	"os"
	"strings"
)

func (r *reject) nameEditor(fi dirreader.FileInfo) error {
	path := strings.Trim(fi.PathAbs, fi.Name())
	name := fi.Name()

	for _, str := range r.delete {
		name = strings.Replace(name, str, "", 1)
	}

	for _, str := range r.space {
		name = strings.Replace(name, str, " ", len(name))
	}

	dst := strings.Join([]string{path, name}, "")

	if err := os.Rename(fi.PathAbs, dst); err != nil {
		return fmt.Errorf("can't rename file %v: %v", fi.PathAbs, err)
	}

	return nil
}
