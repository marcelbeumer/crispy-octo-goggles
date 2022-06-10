package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/alecthomas/kong"
	"github.com/marcelbeumer/go-playground/streamproc/services/consumer-low/internal/log"

	"encoding/json"
	"math/big"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
)

type CLI struct {
	Host       string  `help:"API host."                            short:"h" default:"127.0.0.1" env:"HOST"`
	Port       int     `help:"API port."                            short:"p" default:"9997"      env:"PORT"`
	KafkaHost  string  `help:"Kafka host."                                    default:"127.0.0.1" env:"KAFKA_HOST"`
	KafkaPort  int     `help:"Kafka port."                                    default:"9092"      env:"KAFKA_PORT"`
	KafkaTopic string  `help:"Kafka topic."                                   default:"events"    env:"KAFKA_TOPIC"`
	Watermark  float64 `help:"Number until what is considered low."           default:"5.0"       env:"WATERMARK"`
}

type Event struct {
	Time   int64      `json:"time"`
	Amount *big.Float `json:"amount"`
}

type ResponseBody struct {
	Message string `json:"message"`
}

type Server struct {
	logger    log.Logger
	kafkaConn *kafka.Conn
	ctx       context.Context
	cancel    context.CancelFunc
	watermark float64
}

func NewServer(logger log.Logger, watermark float64) *Server {
	return &Server{
		logger:    logger,
		watermark: watermark,
	}
}

func (s *Server) ListenAndServe(addr string, kafkaAddr string, kafkaTopic string) error {
	logger := s.logger
	logger.Infow("starting server", "addr", addr)

	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cancel

	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaAddr, kafkaTopic, 0)
	if err != nil {
		return err
	}
	s.kafkaConn = conn

	go s.consumeKafka(ctx)

	r := mux.NewRouter()
	r.HandleFunc("/", s.readyProbe).Methods("GET")

	return http.ListenAndServe(addr, r)
}

func (s *Server) Shutdown(ctx context.Context) {
	logger := s.logger
	s.cancel()

	if err := s.kafkaConn.Close(); err != nil {
		logger.Errorw("failed to close kafka connection", log.Error(err))
	}
}

func (s *Server) readyProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}

func (s *Server) consumeKafka(ctx context.Context) {
	logger := s.logger
	for {
		select {
		case <-ctx.Done():
		default:
			conn := s.kafkaConn
			batch := conn.ReadBatch(100, 1e6)
			b := make([]byte, 10e3)
			for {
				n, err := batch.Read(b)
				if err != nil {
					break
				}
				var event Event
				if err := json.Unmarshal(b[:n], &event); err != nil {
					logger.Errorw("failed to parse message", log.Error(err))
					continue
				}

				amount, _ := event.Amount.Float64()
				if amount <= s.watermark {
					logger.Infow("have message with low amount",
						"time", event.Time,
						"amount", event.Amount.String(),
					)
				}
			}

			if err := batch.Close(); err != nil {
				logger.Errorw("failed to close kafka batch", log.Error(err))
			}

		}
	}
}

func (s *Server) writeBadRequest(err error, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(ResponseBody{
		Message: err.Error(),
	})
}

func main() {
	cli := CLI{}
	_ = kong.Parse(
		&cli,
		kong.Name("consumer-low"),
		kong.UsageOnError(),
	)

	zl := log.NewZapLogger(os.Stderr)
	log.RedirectStdLog(zl)
	logger := log.NewZapLoggerAdapter(zl)

	srv := NewServer(logger, cli.Watermark)
	go func() {
		addr := fmt.Sprintf("%s:%d", cli.Host, cli.Port)
		kafkaAddr := fmt.Sprintf("%s:%d", cli.KafkaHost, cli.KafkaPort)
		if err := srv.ListenAndServe(addr, kafkaAddr, cli.KafkaTopic); err != nil {
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
