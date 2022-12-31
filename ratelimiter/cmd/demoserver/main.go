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
	var rate float64
	var burst int

	flag.StringVar(&host, "host", "", "host to listen on")
	flag.IntVar(&port, "port", 8080, "port to listen on")
	flag.Float64Var(&rate, "rate", 1, "rate limit in requests per second")
	flag.IntVar(&burst, "burst", 1, "burst size")
	flag.Parse()

	xtimeManager := ratelimiter.NewHTTPManager(context.Background(), rate, burst, ratelimiter.XtimeLimiterFactory)
	uberManager := ratelimiter.NewHTTPManager(context.Background(), rate, burst, ratelimiter.UberLimiterFactory)

	http.Handle("/xtime", xtimeManager.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world! (x/time/rate)")
	}))

	http.Handle("/uber", uberManager.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world! (x/time/rate)")
	}))

	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Starting server on %s\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
