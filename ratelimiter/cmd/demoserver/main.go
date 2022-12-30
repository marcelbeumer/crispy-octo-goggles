package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

func main() {
	var host string
	var port int

	flag.StringVar(&host, "host", "", "host to listen on")
	flag.IntVar(&port, "port", 8080, "port to listen on")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world!")
	})

	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Starting server on %s\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
