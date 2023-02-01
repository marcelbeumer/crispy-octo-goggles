package server

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type Message struct {
	Message string `json:"message"`
}

type ServerOpts struct {
	SleepSec int
}

type Server struct {
	Opts ServerOpts
	Mux  *http.ServeMux
}

func (s *Server) writeJSON(obj any, statusCode int, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(obj)
}

func (s *Server) write404(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not found"))
}

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.String() != "/" {
		s.write404(w)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, go ahead and POST JSON to /json."))
}

func (s *Server) HandleJSONRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeJSON(
			Message{"method not allowed, please POST instead."},
			http.StatusMethodNotAllowed,
			w)
		return
	}

	contentType := r.Header.Get("content-type")
	if contentType != "application/json" {
		s.writeJSON(
			Message{"wrong content type, please use application/json"},
			http.StatusBadRequest,
			w)
		return
	}

	var data any
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		s.writeJSON(
			Message{fmt.Sprintf("could not parse json: %s", err.Error())},
			http.StatusBadRequest,
			w)
		return
	}

	if s.Opts.SleepSec > 0 {
		r := rand.Intn(s.Opts.SleepSec)
		time.Sleep(time.Duration(r) * time.Second)
	}

	s.writeJSON(data, http.StatusOK, w)
}

func NewServer(opts ServerOpts) *Server {
	mux := http.NewServeMux()
	s := &Server{Mux: mux, Opts: opts}

	mux.HandleFunc("/", s.HandleIndex)
	mux.HandleFunc("/json", s.HandleJSONRequest)

	return s
}
