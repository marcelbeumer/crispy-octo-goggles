package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/gorilla/mux"
	"github.com/marcelbeumer/go-playground/streamproc/services/event-api/internal/log"
	"github.com/segmentio/kafka-go"
)

type ServerOpts struct {
	Host       string `help:"API host."               short:"h" default:"127.0.0.1" env:"HOST"`
	Port       int    `help:"API port."               short:"p" default:"9998"      env:"PORT"`
	KafkaHost  string `help:"Kafka host."                       default:"127.0.0.1" env:"KAFKA_HOST"`
	KafkaPort  int    `help:"Kafka port."                       default:"9092"      env:"KAFKA_PORT"`
	KafkaTopic string `help:"Kafka topic."                      default:"events"    env:"KAFKA_TOPIC"`
	Disable    bool   `help:"Disable all processing." short:"x"                     env:"DISABLE"`
}

type Event struct {
	Time   int64      `json:"time"`
	Amount *big.Float `json:"amount"`
}

type PostEventsJsonBody []Event

type ResponseBody struct {
	Message string `json:"message"`
}

type Server struct {
	logger log.Logger
	opts   ServerOpts
	writer *kafka.Writer
	ctx    context.Context
	cancel context.CancelFunc
}

func NewServer(logger log.Logger, opts ServerOpts) *Server {
	return &Server{
		logger: logger,
		opts:   opts,
	}
}

func (s *Server) ListenAndServe() error {
	opts := s.opts
	logger := s.logger
	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	kafkaAddr := fmt.Sprintf("%s:%d", opts.KafkaHost, opts.KafkaPort)

	logger.Infow("starting server", "addr", addr)

	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cancel
	s.writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{kafkaAddr},
		Topic:    opts.KafkaTopic,
		Balancer: &kafka.LeastBytes{},
	})

	r := mux.NewRouter()
	r.HandleFunc("/", s.readyProbe).Methods("GET")
	r.HandleFunc("/", s.postEvents).Methods("POST")

	return http.ListenAndServe(addr, r)
}

func (s *Server) Shutdown(ctx context.Context) {
	logger := s.logger
	if err := s.writer.Close(); err != nil {
		logger.Errorw("failed to close kafka writer", log.Error(err))
	}
}

func (s *Server) readyProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}

func (s *Server) writeBadRequest(
	err error,
	w http.ResponseWriter,
	r *http.Request,
) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(ResponseBody{
		Message: err.Error(),
	})
}

func (s *Server) postEvents(w http.ResponseWriter, r *http.Request) {
	logger := s.logger

	if s.opts.Disable {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ResponseBody{
			Message: "processing disabled, did nothing",
		})
		return
	}

	var messages []kafka.Message
	contentType := r.Header.Get("Content-Type")

	switch contentType {

	// Validates but wastes time with parsing/serializing JSON
	case "application/json":
		var postEventsBody PostEventsJsonBody
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&postEventsBody)
		if err != nil {
			s.writeBadRequest(err, w, r)
			return
		}

		messages = make([]kafka.Message, len(postEventsBody))
		for i, event := range postEventsBody {
			msg, err := json.Marshal(event)
			if err != nil {
				s.writeBadRequest(err, w, r)
				return
			}
			messages[i] = kafka.Message{Value: msg}
		}

	// Does not validate but passes on messages quickly
	case "application/text":
		scanner := bufio.NewScanner(r.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				messages = append(messages, kafka.Message{
					Value: []byte(line),
				})
			}
		}
		if scanner.Err() != nil {
			s.writeBadRequest(scanner.Err(), w, r)
			return
		}

	default:
		msg := fmt.Errorf("unsupported content type %s", contentType)
		s.writeBadRequest(msg, w, r)
		return
	}

	err := s.writer.WriteMessages(s.ctx, messages...)
	if err != nil {
		s.writeBadRequest(err, w, r)
		return
	}

	logger.Infow("ingested events",
		"count", len(messages),
		"contentType", contentType)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseBody{
		Message: fmt.Sprintf("ingested %d events", len(messages)),
	})
	return
}

func main() {
	opts := ServerOpts{}
	_ = kong.Parse(
		&opts,
		kong.Name("event-api"),
		kong.UsageOnError(),
	)

	zl := log.NewZapLogger(os.Stderr)
	log.RedirectStdLog(zl)
	logger := log.NewZapLoggerAdapter(zl)

	srv := NewServer(logger, opts)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Errorw("could not start server", log.Error(err))
			os.Exit(1)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	srv.Shutdown(ctx)
	os.Exit(0)
}
