package reject

import (
	"github.com/gromey/filemanager/common"
	"github.com/gromey/filemanager/dirreader"
	"log"
)

type reject struct {
	paths     []string
	extension []string
	include   bool
	details   bool
	delete    []string
	space     []string
}

func New(c *Config) *reject {
	r := &reject{
		paths:  c.Paths,
		delete: c.Delete,
		space:  c.Space,
	}

	if c.Mask.On {
		r.extension = c.Mask.Extension
		r.include = c.Mask.Include
		r.details = c.Mask.Details
	}

	return r
}

func (r *reject) Start() error {
	var excl, incl []dirreader.FileInfo

	for _, path := range r.paths {
		ex, in, err := dirreader.SetDirReader(path, r.extension, r.include, r.details, false).Exec()
		if err != nil {
			return err
		}

		excl = append(excl, ex...)
		incl = append(incl, in...)
	}

	if r.details {
		log.Printf("%d %s\n", len(excl), "files was excluded by mask.")
	}

	for _, fi := range incl {
		if err := r.nameEditor(fi); err != nil {
			return err
		}
	}

	return nil
}

func Run(config string) error {
	var c []*Config

	err := common.GetConfig(config, c)
	if err != nil {
		return err
	}

	for _, cfg := range c {
		err = New(cfg).Start()
		if err != nil {
			return err
		}
	}

	log.Println("Editing completed!")

	return nil
}
