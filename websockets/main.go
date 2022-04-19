package main

import (
	"fmt"

	"github.com/alecthomas/kong"
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
		kong.Name("websockets"),
		kong.UsageOnError(),
	)
	switch ctx.Command() {
	case "client":
		fmt.Println(ctx.Command())
	case "server":
		fmt.Println(cli.Server.Host, cli.Server.Port)
	}
}
