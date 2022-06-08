package server

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/marcelbeumer/go-playground/streamproc/services/event-api/internal/log"
)

type Event struct {
	Time   int64     `json:"time"`
	Amount big.Float `json:"amount"`
}

type PostEventsBody []Event

type ResponseBody struct {
	Message string `json:"message"`
}

type Server struct {
	logger log.Logger
}

func NewServer(logger log.Logger) *Server {
	return &Server{
		logger: logger,
	}
}

func (s *Server) ListenAndServe(addr string) error {
	logger := s.logger
	logger.Infow("starting server", "addr", addr)

	r := mux.NewRouter()
	r.HandleFunc("/", s.postEvents).Methods("POST")

	err := http.ListenAndServe(addr, r)
	return err
}

func (s *Server) Shutdown(ctx context.Context) {
	//
}

func (s *Server) postEvents(w http.ResponseWriter, r *http.Request) {
	var postEventsBody PostEventsBody
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&postEventsBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseBody{
			Message: err.Error(),
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseBody{
		Message: fmt.Sprintf("ingested %d events", len(postEventsBody)),
	})
	return
}
