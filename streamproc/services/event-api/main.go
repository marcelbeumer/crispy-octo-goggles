package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/alecthomas/kong"
	"github.com/marcelbeumer/go-playground/streamproc/services/event-api/internal/log"
	"github.com/marcelbeumer/go-playground/streamproc/services/event-api/internal/server"
)

type CLI struct {
	Host string `help:"API host." short:"h" default:"127.0.0.1" env:"HOST"`
	Port int    `help:"API port." short:"p" default:"9998"      env:"PORT"`
}

func main() {
	cli := CLI{}
	_ = kong.Parse(
		&cli,
		kong.Name("event-api"),
		kong.UsageOnError(),
	)

	zl := log.NewZapLogger(os.Stderr)
	log.RedirectStdLog(zl)
	logger := log.NewZapLoggerAdapter(zl)

	srv := server.NewServer(logger)
	go func() {
		addr := fmt.Sprintf("%s:%d", cli.Host, cli.Port)
		if err := srv.ListenAndServe(addr); err != nil {
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
