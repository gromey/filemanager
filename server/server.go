package server

import (
	_ "embed"
	"github.com/gromey/filemanager/duplicate"
	"html/template"
	"net/http"
)

type server struct {
	cfg *Config
}

// New configures and creates a new server.
func New(cfg *Config) *server {
	return &server{
		cfg: cfg,
	}
}

// Start starts the server.
func (s *server) Start() error {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("/", s.home())
	mux.HandleFunc("/synchronizing", s.synchronizing())
	mux.HandleFunc("/duplicate", s.duplicate())

	return http.ListenAndServe(":8888", mux)
}

type Button struct {
	Path string
	Name string
}

type Data struct {
	Page          string
	Buttons       []Button
	WorkContainer template.HTML
	Items         []string
	Dubl          []duplicate.Test
	TD            int
	TDF           int
}
