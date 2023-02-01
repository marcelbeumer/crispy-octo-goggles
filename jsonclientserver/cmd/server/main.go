package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/marcelbeumer/go-playground/jsonclientserver/internal/server"
)

func exit(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}

func main() {
	var host string
	var port int
	var sleepSec int

	flag.StringVar(&host, "host", "", "host to listen on")
	flag.IntVar(&port, "port", 8080, "port to listen on")
	flag.IntVar(&sleepSec, "sleep", 1, "sleep 0-<sleep> seconds before http response (0=off)")
	flag.Parse()

	svr := server.NewServer(server.ServerOpts{
		SleepSec: sleepSec,
	})
	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Starting server on %s\n", addr)

	if err := http.ListenAndServe(addr, svr.Mux); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
