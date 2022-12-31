package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/marcelbeumer/go-playground/ratelimiter"
)

func main() {
	var host string
	var port int

	flag.StringVar(&host, "host", "", "host to listen on")
	flag.IntVar(&port, "port", 8080, "port to listen on")
	flag.Parse()

	xtimeManager := ratelimiter.NewHTTPManager(context.Background(), 1, 1, ratelimiter.XtimeLimiterFactory)

	http.Handle("/xtime", xtimeManager.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world! (x/time/rate)")
	}))

	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Starting server on %s\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
