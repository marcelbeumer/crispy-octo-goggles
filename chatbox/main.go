package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/client"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/server"
)

type ClientServerOpts struct {
	Host string `help:"Server host." short:"h" default:"127.0.0.1"`
	Port int    `help:"Server port." short:"p" default:"9998"`
}

type Commands struct {
	Client struct {
		ClientServerOpts
	} `cmd:"client" help:"Start client"`
	Server struct {
		ClientServerOpts
	} `cmd:"client" help:"Start server"`
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
		if err := client.StartClient(addr); err != nil {
			panic(err)
		}
	case "server":
		addr := fmt.Sprintf("%s:%d", cli.Server.Host, cli.Server.Port)
		if err := server.StartServer(addr); err != nil {
			panic(err)
		}
	}
}
