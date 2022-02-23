package duplicate

import (
	"github.com/gromey/filemanager/common"
	"github.com/gromey/filemanager/dirreader"
	"log"
)

type duplicate struct {
	paths     []string
	extension []string
	include   bool
	details   bool
}

func New(c *Config) *duplicate {
	d := &duplicate{
		paths: c.Paths,
	}

	if c.Mask.On {
		d.extension = c.Mask.Extension
		d.include = c.Mask.Include
		d.details = c.Mask.Details
	}

	return d
}

func (d *duplicate) Start() ([]Repeated, error) {
	var excluded, included []dirreader.FileInfo

	for _, path := range d.paths {
		exclude, include, err := dirreader.SetDirReader(path, d.extension, d.include, d.details, true).Exec()
		if err != nil {
			return nil, err
		}

		excluded = append(excluded, exclude...)
		included = append(included, include...)
	}

	if d.details {
		log.Printf("%d files was excluded by mask.\n", len(excluded))
		//for _, v := range excluded {
		//	log.Println(v)
		//}
	}

	match := compare(included)
	if len(match) == 0 {
		return nil, nil
	}

	return match, nil
}

func Run(config string) error {
	var c []*Config

	err := common.GetConfig(config, &c)
	if err != nil {
		return err
	}

	for _, cfg := range c {
		var res []Repeated
		res, err = New(cfg).Start()
		if err != nil {
			return err
		}

		for _, v := range res {
			log.Println(v)
		}
	}

	return nil
}
