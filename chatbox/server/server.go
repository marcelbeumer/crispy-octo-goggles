package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct{}

func (s *Server) Start(addr string) error {
	r := s.initRouting()
	err := http.ListenAndServe(addr, r)
	return err
}

func (s *Server) Stop() error {
	return nil
}

func (s *Server) initRouting() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", s.handleIndex)
	return r
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
}

func NewServer() *Server {
	s := Server{}
	return &s
}
