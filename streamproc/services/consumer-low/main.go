package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/alecthomas/kong"
	"github.com/marcelbeumer/go-playground/streamproc/services/consumer-high/internal/log"

	"encoding/json"
	"math/big"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
)

type ServerOpts struct {
	Watermark  float64 `help:"Number until what is considered low." default:"5.0"       env:"WATERMARK"`
	Host       string  `help:"API host."                            default:"127.0.0.1" env:"HOST"        short:"h"`
	Port       int     `help:"API port."                            default:"9996"      env:"PORT"        short:"p"`
	KafkaHost  string  `help:"Kafka host."                          default:"127.0.0.1" env:"KAFKA_HOST"`
	KafkaPort  int     `help:"Kafka port."                          default:"9092"      env:"KAFKA_PORT"`
	KafkaTopic string  `help:"Kafka topic."                         default:"events"    env:"KAFKA_TOPIC"`
	Debug      bool    `help:"Show debug messages."                                                       short:"d"`
}

type Event struct {
	Time   int64      `json:"time"`
	Amount *big.Float `json:"amount"`
}

type ResponseBody struct {
	Message string `json:"message"`
}

type EventBuffer struct {
	events []Event
	l      sync.RWMutex
	offset int64
}

func (b *EventBuffer) Len() int {
	return len(b.events)
}

func (b *EventBuffer) Append(e Event, offset int64) {
	b.l.Lock()
	defer b.l.Unlock()
	b.events = append(b.events, e)
	b.offset = offset
}

func (b *EventBuffer) Recover(e []Event) {
	b.l.Lock()
	defer b.l.Unlock()
	b.events = append(e, b.events...)
}

func (b *EventBuffer) Flush() ([]Event, int64) {
	b.l.Lock()
	defer b.l.Unlock()
	slice := b.events[:]
	b.events = make([]Event, 0)
	return slice, b.offset
}

type Server struct {
	opts        ServerOpts
	startOffset int64
	buf         EventBuffer
	logger      log.Logger
	reader      *kafka.Reader
	ctx         context.Context
	cancel      context.CancelFunc
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
	logger.Infow("starting server", "addr", addr)

	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cancel

	// if err := s.setupInfluxDB(); err != nil {
	// 	return err
	// }
	// logger.Infow("influxDB set up")

	if err := s.setupKafka(); err != nil {
		return err
	}
	logger.Infow("kafka set up",
		"startOffset", s.startOffset)

	go s.pumpKafka(ctx)
	go s.pumpInfluxDb(ctx)

	r := mux.NewRouter()
	r.HandleFunc("/", s.readyProbe).Methods("GET")

	logger.Infow("listening", "addr", addr)
	return http.ListenAndServe(addr, r)
}

func (s *Server) Shutdown(ctx context.Context) {
	logger := s.logger
	s.cancel()

	if s.reader != nil {
		if err := s.reader.Close(); err != nil {
			logger.Errorw("failed to close kafka reader", log.Error(err))
		}
	}
}

func (s *Server) readyProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}

func (s *Server) setupKafka() error {
	kafkaAddr := fmt.Sprintf("%s:%d", s.opts.KafkaHost, s.opts.KafkaPort)
	s.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{kafkaAddr},
		Topic:     s.opts.KafkaTopic,
		Partition: 0,
		MinBytes:  10,
		MaxBytes:  10e6, // 10MB
	})
	s.reader.SetOffset(s.startOffset)
	return nil
}

func (s *Server) handleStorageError(err error, msg string) {
	logger := s.logger
	sleepSec := 10
	logger.Errorw(
		msg,
		log.Error(err),
		"sleepSec", sleepSec,
	)
	time.Sleep(time.Second * time.Duration(sleepSec))
}

func (s *Server) pumpKafka(ctx context.Context) {
	logger := s.logger

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if s.buf.Len() >= 10000 {
			logger.Warn("event buffer too full, waiting with reading from kafka")
			time.Sleep(time.Second * 2)
			continue
		}

		m, err := s.reader.ReadMessage(context.Background())
		if err != nil {
			s.handleStorageError(err, "failed to read message (sleeping before retrying)")
			continue
		}
		var event Event
		if err := json.Unmarshal(m.Value, &event); err != nil {
			s.handleStorageError(err, "failed to parse message")
			continue
		}

		amount, _ := event.Amount.Float64()
		if amount <= s.opts.Watermark {
			logger.Debugw("have message with low amount",
				"time", event.Time,
				"amount", event.Amount.String(),
			)
			s.buf.Append(event, s.reader.Offset())
		}
	}
}

func (s *Server) pumpInfluxDb(ctx context.Context) {
	logger := s.logger
	for {
		select {
		case <-ctx.Done():
		default:
		}

		events, offset := s.buf.Flush()
		count := len(events)
		if count == 0 {
			time.Sleep(time.Second * 2)
			continue
		}
		_ = logger
		_ = offset
	}
}

func (s *Server) writeBadRequest(err error, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(ResponseBody{
		Message: err.Error(),
	})
}

func main() {
	opts := ServerOpts{}
	_ = kong.Parse(
		&opts,
		kong.Name("consumer-low"),
		kong.UsageOnError(),
	)

	zl := log.NewZapLogger(os.Stderr, opts.Debug)
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
