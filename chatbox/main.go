package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/websocket"
)

type ClientServerOpts struct {
	Host string `help:"Server host." short:"h" default:"127.0.0.1"`
	Port int    `help:"Server port." short:"p" default:"9998"`
}
type ClientOpts struct {
	Username string `help:"Username." required:"" short:"u"`
}

type Commands struct {
	Client struct {
		ClientServerOpts
		ClientOpts
	} `cmd:"client" help:"Start client"`
	Server struct {
		ClientServerOpts
	} `cmd:"client" help:"Start server"`
	Test struct {
	} `cmd:"test"   help:"Run non-http test scenario"`
}

func main() {
	cli := Commands{}
	ctx := kong.Parse(
		&cli,
		kong.Name("chatbox"),
		kong.UsageOnError(),
	)
	switch ctx.Command() {
	case "client":
		addr := fmt.Sprintf("%s:%d", cli.Server.Host, cli.Server.Port)
		if err := websocket.StartClient(addr, cli.Client.Username); err != nil {
			panic(err)
		}
	case "server":
		addr := fmt.Sprintf("%s:%d", cli.Server.Host, cli.Server.Port)
		if _, err := websocket.StartServer(addr); err != nil {
			panic(err)
		}
	case "test":
		fmt.Println("test")
	}
}
