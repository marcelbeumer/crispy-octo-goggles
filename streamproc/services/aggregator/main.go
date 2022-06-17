package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"sync"
	"time"

	"github.com/alecthomas/kong"
	"github.com/marcelbeumer/go-playground/streamproc/services/aggregator/internal/log"

	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"

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

type EventType = string

const (
	EventTypeLow  EventType = "low"
	EventTypeHigh EventType = "high"
)

type DataPoint struct {
	Time  int64     `json:"time"`
	Count int64     `json:"count"`
	Type  EventType `json:"type"`
}

type ResponseBody struct {
	Message string      `json:"message"`
	Points  []DataPoint `json:"points"`
}

type Server struct {
	opts     ServerOpts
	logger   log.Logger
	influxQ  influxdbApi.QueryAPI
	tsdbPool *pgxpool.Pool
	ctx      context.Context
	cancel   context.CancelFunc
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

	if err := s.setupTSDB(); err != nil {
		return err
	}
	logger.Infow("TSDB set up")

	if err := s.setupInfluxDB(); err != nil {
		return err
	}
	logger.Infow("influxDB set up")

	r := mux.NewRouter()
	r.HandleFunc("/", s.readyProbe).Methods("GET")
	r.HandleFunc("/data", s.getData).Methods("GET")

	logger.Infow("listening", "addr", addr)
	return http.ListenAndServe(addr, r)
}

func (s *Server) Shutdown(ctx context.Context) {
	s.cancel()
	s.tsdbPool.Close()
}

func (s *Server) readyProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}

func (s *Server) getData(w http.ResponseWriter, r *http.Request) {
	logger := s.logger

	if s.opts.Disable {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ResponseBody{
			Message: "processing disabled, did nothing",
		})
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	total := "10m"
	step := "10s"

	var low []DataPoint
	go func() {
		defer wg.Done()

		query := fmt.Sprintf(`
			from(bucket: "%s")
				|> range(start: %s)
				|> filter(fn: (r) => r._measurement == "event" and r._field == "amount")
				|> aggregateWindow(every: %s, fn: count)
		`, s.opts.InfluxDBBucket, "-"+total, step)

		res, err := s.influxQ.Query(s.ctx, query)
		if err != nil {
			logger.Error("could not get events from influx", log.Error(err))
		}

		for res.Next() {
			record := res.Record()
			value, ok := record.Value().(int64)
			if !ok {
				logger.Error("offset has wrong data type",
					"value", record.Value())
				continue
			}

			low = append(low, DataPoint{
				Time:  record.Time().UnixMilli(),
				Count: value,
				Type:  EventTypeLow,
			})
		}

	}()

	var high []DataPoint
	go func() {
		defer wg.Done()
		query := fmt.Sprintf(`
			SELECT time_bucket('%s', "time") AS bucket, count(amount) FROM events
			WHERE time > now() - INTERVAL '%s'
			GROUP BY bucket
		`, step, total)
		rows, err := s.tsdbPool.Query(s.ctx, query)
		if err != nil {
			logger.Error("could not get events from tsdb", log.Error(err))
			return
		}
		defer rows.Close()
		for rows.Next() {
			var t time.Time
			var v int64
			if err := rows.Scan(&t, &v); err != nil {
				logger.Error("could not scan tsdb row", log.Error(err))
				return // give up right away; prevent many more errors
			}
			high = append(high, DataPoint{
				Time:  t.UnixMilli(),
				Count: v,
				Type:  EventTypeHigh,
			})
		}
	}()

	wg.Wait()

	all := make([]DataPoint, 0, len(low)+len(high))
	all = append(all, low...)
	all = append(all, high...)
	sort.Slice(all, func(a, b int) bool {
		return all[a].Time < all[b].Time
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseBody{
		Message: "ok",
		Points:  all,
	})
}

func (s *Server) setupTSDB() error {
	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		default:
		}

		sslmode := "disable"
		if s.opts.PostgresUseSSL == "1" {
			sslmode = "require"
		}
		pgAddr := fmt.Sprintf(
			"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
			s.opts.PostgresUser,
			s.opts.PostgresPassword,
			s.opts.PostgresHost,
			s.opts.PostgresPort,
			s.opts.PostgresDb,
			sslmode,
		)
		conn, err := pgxpool.Connect(s.ctx, pgAddr)
		if err != nil {
			msg := "could not connect to database (will retry)"
			s.logger.Errorw(msg, log.Error(err))
			time.Sleep(time.Millisecond * 2000)
			continue
		}
		s.tsdbPool = conn
		break
	}

	return nil
}

func (s *Server) setupInfluxDB() error {
	url := fmt.Sprintf("http://%s", s.opts.InfluxDBHost)
	client := influxdb2.NewClient(url, s.opts.InfluxDBToken)
	s.influxQ = client.QueryAPI(s.opts.InfluxDBOrg)
	return nil
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
