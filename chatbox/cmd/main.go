package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/log"
)

type ClientServerOpts struct {
	Host string `help:"Server host." short:"h" default:"127.0.0.1"`
	Port int    `help:"Server port." short:"p" default:"9998"`
}

type ClientOpts struct {
	Username string `help:"Username." required:"" short:"u"`
}

type Commands struct {
	Verbose     bool `help:"Verbose (logging info)"       short:"v"`
	VeryVerbose bool `help:"Very verbose (logging debug)" short:"V"`

	Client struct {
		ClientServerOpts
		ClientOpts
	} `help:"Start client"               cmd:"client"`
	Server struct {
		ClientServerOpts
	} `help:"Start server"               cmd:"client"`
	Test struct {
	} `help:"Run non-http test scenario" cmd:"test"`
}

func main() {
	cli := Commands{}
	ctx := kong.Parse(
		&cli,
		kong.Name("chatbox"),
		kong.UsageOnError(),
	)

	logger := log.NewDefaultLogger(cli.Verbose, cli.VeryVerbose)
	log.SetStandardLogger(logger)

	switch ctx.Command() {
	case "client":
		addr := fmt.Sprintf("%s:%d", cli.Server.Host, cli.Server.Port)
		conn, err := chatbox.NewWebsocketClientConnection(addr, cli.Client.Username, logger)
		if err != nil {
			panic(err)
		}
		defer conn.Close()
		// fe := chatbox.NewStdoutFrontend(conn, logger)
		// if err := fe.Start(); err != nil {
		// 	panic(err)
		// }
		fe, err := chatbox.NewGUIFrontend(conn, logger)
		if err != nil {
			panic(err)
		}
		if err := fe.Start(); err != nil {
			panic(err)
		}

	case "server":
		addr := fmt.Sprintf("%s:%d", cli.Server.Host, cli.Server.Port)
		s := chatbox.NewWebsocketServer(logger)
		if err := s.Start(addr); err != nil {
			panic(err)
		}

	case "test":
		hub := chatbox.NewHub(logger)
		toHubCh := make(chan chatbox.Event)
		toUserCh := make(chan chatbox.Event)
		hubConn := chatbox.NewTestConnection(toHubCh, toUserCh)
		userConn := chatbox.NewTestConnection(toUserCh, toHubCh)
		defer hubConn.Close()
		defer userConn.Close()
		if err := hub.ConnectUser("User", hubConn); err != nil {
			panic(err)
		}
		fe := chatbox.NewStdoutFrontend(userConn, logger)
		if err := fe.Start(); err != nil {
			panic(err)
		}
	}
}
