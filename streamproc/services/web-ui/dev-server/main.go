package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/alecthomas/kong"
	"github.com/gorilla/mux"
)

type Opts struct {
	BackendUrl string `help:"Backend url." env:"BACKEND_URL" short:"b" required:""`
	Host       string `help:"Host."        env:"HOST"        short:"h"             default:"127.0.0.1"`
	Port       int    `help:"Port."        env:"PORT"        short:"p"             default:"8080"`
	StaticDir  string `help:"Static dir."  env:"STATIC_DIR"  short:"s"             default:"../public"`
}

func start(opts Opts) error {
	dir := "../public"

	url, err := url.Parse(opts.BackendUrl)
	if err != nil {
		return err
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	fmt.Printf("/api proxy to %s\n", opts.BackendUrl)

	r := mux.NewRouter()
	r.PathPrefix("/api").Handler(http.StripPrefix("/api", proxy))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(dir)))

	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	fmt.Printf("starting dev server on %s\n", addr)
	return http.ListenAndServe(addr, r)
}

func main() {
	opts := Opts{}
	_ = kong.Parse(
		&opts,
		kong.Name("dev-server"),
		kong.UsageOnError(),
	)

	if err := start(opts); err != nil {
		log.Fatal(err)
	}
}
