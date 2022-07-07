package main

import (
	"fmt"

	"github.com/alecthomas/kong"
)

type Opts struct {
	Show struct{} `help:"Show package info"       cmd:"show"`
	List struct{} `help:"List available packages" cmd:"list"`
}

func main() {
	opts := Opts{}
	ctx := kong.Parse(
		&opts,
		kong.Name("packex"),
		kong.UsageOnError(),
	)
	switch ctx.Command() {
	case "show":
		fmt.Println("show")
	case "list":
		fmt.Println("list")
	}
}
