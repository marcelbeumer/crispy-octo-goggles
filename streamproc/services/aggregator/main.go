package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/alecthomas/kong"
	"github.com/marcelbeumer/go-playground/streamproc/services/aggregator/internal/log"

	"math/big"
	"net/http"

	"github.com/gorilla/mux"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdbApi "github.com/influxdata/influxdb-client-go/v2/api"
)

type ServerOpts struct {
	Watermark        float64 `help:"Number until what is considered low." default:"5.0"       env:"WATERMARK"`
	Host             string  `help:"API host."                            default:"127.0.0.1" env:"HOST"              short:"h"`
	Port             int     `help:"API port."                            default:"9996"      env:"PORT"              short:"p"`
	PostgresHost     string  `help:"Postgres host."                       default:"127.0.0.1" env:"POSTGRES_HOST"`
	PostgresPort     int     `help:"Postgres port."                       default:"5432"      env:"POSTGRES_PORT"`
	PostgresUser     string  `help:"Postgres user."                       default:"user"      env:"POSTGRES_USER"`
	PostgresPassword string  `help:"Postgres password."                   default:"pass"      env:"POSTGRES_PASSWORD"`
	PostgresUseSSL   string  `help:"Postgres use SSL."                    default:"0"         env:"POSTGRES_USE_SSL"`
	PostgresDb       string  `help:"Postgres database."                   default:"postgres"  env:"POSTGRES_DB"`
	InfluxDBHost     string  `help:"InfluxDB host."                                           env:"INFLUXDB_HOST"`
	InfluxDBToken    string  `help:"InfluxDB token."                                          env:"INFLUXDB_TOKEN"`
	InfluxDBOrg      string  `help:"InfluxDB org."                                            env:"INFLUXDB_ORG"`
	InfluxDBBucket   string  `help:"InfluxDB bucket."                                         env:"INFLUXDB_BUCKET"`
	Debug            bool    `help:"Show debug messages."                                                             short:"d"`
	Disable          bool    `help:"Disable all processing."                                  env:"DISABLE"           short:"x"`
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
	writer      influxdbApi.WriteAPIBlocking
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

	// if opts.Disable {
	// 	logger.Info("processing disabled")
	// } else {
	// 	if err := s.setupInfluxDB(); err != nil {
	// 		return err
	// 	}
	// 	logger.Infow("influxDB set up")
	//
	// 	if err := s.setupKafka(); err != nil {
	// 		return err
	// 	}
	// 	logger.Infow("kafka set up",
	// 		"startOffset", s.startOffset)
	//
	// 	go s.pumpKafka(ctx)
	// 	go s.pumpInfluxDb(ctx)
	// }

	r := mux.NewRouter()
	r.HandleFunc("/", s.readyProbe).Methods("GET")

	logger.Infow("listening", "addr", addr)
	return http.ListenAndServe(addr, r)
}

func (s *Server) Shutdown(ctx context.Context) {
	s.cancel()
}

func (s *Server) readyProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}

func (s *Server) setupInfluxDB() error {
	logger := s.logger
	url := fmt.Sprintf("http://%s", s.opts.InfluxDBHost)
	client := influxdb2.NewClient(url, s.opts.InfluxDBToken)
	s.writer = client.WriteAPIBlocking(
		s.opts.InfluxDBOrg,
		s.opts.InfluxDBBucket,
	)

	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		default:
		}

		q := client.QueryAPI(s.opts.InfluxDBOrg)

		query := fmt.Sprintf(`
			from(bucket: "%s")
				|> range(start: -1h)
				|> filter(fn: (r) => r._measurement == "event" and r._field == "offset")
				|> last()
		`, s.opts.InfluxDBBucket)
		res, err := q.Query(s.ctx, query)

		if err != nil {
			msg := "could not get last offset (will retry)"
			logger.Errorw(msg, log.Error(err))
			time.Sleep(time.Millisecond * 2000)
			continue
		}

		for res.Next() {
			record := res.Record()
			value, ok := record.Value().(int64)
			if !ok {
				return fmt.Errorf(
					"offset has wrong data type (value: %v)",
					record.Value(),
				)
			}
			s.startOffset = value
		}
		break
	}

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

func main() {
	opts := ServerOpts{}
	_ = kong.Parse(
		&opts,
		kong.Name("aggregator"),
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
