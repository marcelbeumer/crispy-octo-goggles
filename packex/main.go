package main

import (
	"fmt"
	"log"

	"github.com/alecthomas/kong"
	"github.com/marcelbeumer/go-playground/packex/internal/pinfo"
)

type Opts struct {
	Show struct {
		Name string `arg:""`
	} `help:"Show package info"       cmd:"show"`
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
	case "show <name>":
		pkg := pinfo.New(opts.Show.Name)
		if err := pkg.Resolve(); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", pkg)
	case "list":
		fmt.Println("list")
	}
}
