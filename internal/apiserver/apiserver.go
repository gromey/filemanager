package apiserver

import (
	logger2 "github.com/GroM1124/filemanager/pkg/logger"
	"github.com/gorilla/mux"
	"io"
	"net/http"
)

// APIServer ...
type APIServer struct {
	cfg    *Config
	lgr    logger2.Logger
	router *mux.Router
}

// New configures and creates a new APIServer.
func New(cfg *Config) *APIServer {
	return &APIServer{
		cfg:    cfg,
		router: mux.NewRouter(),
	}
}

// Start starts the APIServer.
func (s *APIServer) Start() error {
	if err := s.configureLogger(); err != nil {
		return err
	}

	s.configureRouter()

	s.lgr.Info("starting api server on port: ", s.cfg.BindAddr)

	return http.ListenAndServe(s.cfg.BindAddr, s.router)
}

func (s *APIServer) configureLogger() error {
	lgr, err := logger2.New(s.cfg.LogLevel)
	if err != nil {
		return err
	}
	s.lgr = lgr
	return nil
}

// configureRouter configures the APIServer router.
func (s *APIServer) configureRouter() {
	s.router.HandleFunc("/hello", s.hello())
}

// hello ...
func (s *APIServer) hello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "hello")
	}
}
