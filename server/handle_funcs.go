package server

import (
	"bytes"
	"fmt"
	"github.com/gromey/filemanager/duplicate"
	"html/template"
	"log"
	"net/http"
)

// home ...
func (s *server) home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		data := Data{
			Page: "File manager",
			Buttons: []Button{
				{
					Path: "synchronizing",
					Name: "Synchronizing files",
				},
				{
					Path: "duplicate",
					Name: "Find duplicate files",
				},
			},
		}

		base, err := template.ParseFiles("server/template/base.tmpl")
		if err != nil {
			fmt.Fprintf(w, "Unable to load template")
		}

		base.Execute(w, data)
	}
}

// synchronizing ...
func (s *server) synchronizing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		base, err := template.ParseFiles("server/template/base.tmpl")
		if err != nil {
			fmt.Fprintf(w, "Unable to load template")
		}

		workContainer, err := template.ParseFiles("server/template/work_cs_result.tmpl")
		if err != nil {
			fmt.Fprintf(w, "Unable to load template")
		}

		var b bytes.Buffer

		data := Data{
			Page: "Synchronizing files",
			Buttons: []Button{
				{
					Name: "Home",
				},
			},
			Items: []string{"SomeItem1", "SomeItem2", "SomeItem3", "SomeItem4", "SomeItem5", "SomeItem6"},
		}

		workContainer.Execute(&b, data)

		data.WorkContainer = template.HTML(b.String())

		base.Execute(w, data)
	}
}

// duplicate ...
func (s *server) duplicate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		data := Data{
			Page: "Find duplicate files",
			Buttons: []Button{
				{
					Name: "Home",
				},
			},
		}

		base, err := template.ParseFiles("server/template/base.tmpl")
		if err != nil {
			fmt.Fprintf(w, "Unable to load template")
		}

		workContainer, err := template.ParseFiles("server/template/work_cd_result.tmpl")
		if err != nil {
			fmt.Fprintf(w, "Unable to load template")
		}

		var b bytes.Buffer

		fbf, err := duplicate.New(&duplicate.Config{
			Paths: []string{
				"/home/evgeniy/Рабочий стол/Music",
			},
		}).Start()
		if err != nil {
			log.Fatal(err)
		}

		data.Dubl = fbf
		data.TD = len(fbf)

		tdf := 0
		for _, t := range fbf {
			tdf += len(t.Paths)
		}
		data.TDF = tdf

		err = workContainer.Execute(&b, data)
		if err != nil {
			fmt.Fprintf(w, "Ufdebgetrbe")
			return
		}

		data.WorkContainer = template.HTML(b.String())

		base.Execute(w, data)
	}
}
