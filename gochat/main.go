package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/marcelbeumer/go-playground/gochat/internal/chat"
	"github.com/marcelbeumer/go-playground/gochat/internal/grpc"
	"github.com/marcelbeumer/go-playground/gochat/internal/log"
	"github.com/marcelbeumer/go-playground/gochat/internal/websocket"
)

type ClientServerOpts struct {
	Host string `help:"Server host."                   short:"h" default:"127.0.0.1" env:"HOST"`
	Port int    `help:"Server port."                   short:"p" default:"9998"      env:"PORT"`
	Grpc bool   `help:"Use GRPC instead of websockets"`
}

type ClientOpts struct {
	Username       string `help:"Username."                   required:"" short:"u"`
	StdoutFrontend bool   `help:"Use simple stdout frontend."             short:"s"`
}

type Commands struct {
	Verbose     bool `help:"Verbose (logging info)"       short:"v"`
	VeryVerbose bool `help:"Very verbose (logging debug)" short:"V"`
	Client      struct {
		ClientServerOpts
		ClientOpts
	} `help:"Start client"                           cmd:"client"`
	Server struct {
		ClientServerOpts
	} `help:"Start server"                           cmd:"client"`
}

func main() {
	cli := Commands{}
	ctx := kong.Parse(
		&cli,
		kong.Name("gochat"),
		kong.UsageOnError(),
	)

	switch ctx.Command() {

	case "client":
		stdErrBuf := bufio.NewWriter(os.Stderr)
		zl := log.NewZapLogger(stdErrBuf, cli.Verbose, cli.VeryVerbose)
		log.RedirectStdLog(zl)
		logger := log.NewZapLoggerAdapter(zl)

		exit := func(code int) {
			_ = zl.Sync()
			stdErrBuf.Flush()
			os.Exit(code)
		}

		defer exit(0)

		addr := fmt.Sprintf("%s:%d", cli.Server.Host, cli.Server.Port)

		var conn chat.Connection
		var err error

		if cli.Client.Grpc {
			conn, err = grpc.NewClientConnection(addr, cli.Client.Username, logger)
		} else {
			conn, err = websocket.NewClientConnection(addr, cli.Client.Username, logger)
		}

		if err != nil {
			logger.Errorw(
				"could not create connection",
				log.Error(err),
				"addr", addr,
			)
			exit(1)
		}

		defer conn.Close(nil)

		frontendErr := func(err error) {
			logger.Error("frontend error", log.Error(err))
			exit(1)
		}

		if cli.Client.StdoutFrontend {
			fe := chat.NewStdoutFrontend(conn, logger)
			if err := fe.Start(); err != nil {
				frontendErr(err)
			}
		} else {
			fe, err := chat.NewGUIFrontend(conn, logger)
			if err != nil {
				frontendErr(err)
			}
			if err := fe.Start(); err != nil {
				frontendErr(err)
			}
		}

	case "server":
		zl := log.NewZapLogger(os.Stderr, cli.Verbose, cli.VeryVerbose)
		log.RedirectStdLog(zl)
		logger := log.NewZapLoggerAdapter(zl)
		exit := func(code int) {
			_ = zl.Sync()
			os.Exit(code)
		}

		addr := fmt.Sprintf("%s:%d", cli.Server.Host, cli.Server.Port)

		if cli.Server.Grpc {
			s := grpc.NewServer(logger)
			if err := s.Start(addr); err != nil {
				logger.Error("server error", log.Error(err))
				exit(1)
			}

		} else {
			s := websocket.NewServer(logger)
			if err := s.Start(addr); err != nil {
				logger.Error("server error", log.Error(err))
				exit(1)
			}
		}
	}
}
