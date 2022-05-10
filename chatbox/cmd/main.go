package main

import (
	"encoding/json"
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
		e := chatbox.EventNewMessage{
			EventMeta: *chatbox.NewEventMetaNow(),
			Sender:    "User",
			Message:   "Hello",
		}
		m := chatbox.WebsocketMessage{
			Name: chatbox.WEBSOCKET_EVENT_NEW_MESSAGE,
			Data: e,
		}

		b, err := json.Marshal(&m)
		if err != nil {
			panic(err)
		}

		println("name", m.Name)
		println("json", string(b))

		var m2 chatbox.WebsocketMessage
		if err := json.Unmarshal(b, &m2); err != nil {
			panic(err)
		}

		// println(string(b))
	}
}
