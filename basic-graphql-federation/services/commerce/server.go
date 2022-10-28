package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/alecthomas/kong"
	"github.com/marcelbeumer/go-playground/basic-graphql-federation/services/commerce/graph"
	"github.com/marcelbeumer/go-playground/basic-graphql-federation/services/commerce/graph/generated"
)

type ServerOpts struct {
	Host string `help:"API host." default:"127.0.0.1" env:"HOST" short:"h"`
	Port int    `help:"API port." default:"4002"      env:"PORT" short:"p"`
}

func main() {
	opts := ServerOpts{}
	_ = kong.Parse(
		&opts,
		kong.Name("commerce"),
		kong.UsageOnError(),
	)

	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)

	c := generated.Config{Resolvers: graph.NewResolver()}
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(c))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://%s for GraphQL playground", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
