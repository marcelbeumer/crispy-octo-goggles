package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// dev-server --host 0.0.0.0 --port 8080 --proxy /api=http://foo.com/api:8080 --proxy /api2=http://bar.copm/api:9090 --static /=../public

func proxy() error {
	host := "127.0.0.1"
	port := 8888
	dir := "../public"
	backend := "http://kubernetes.docker.internal/api"
	url, err := url.Parse(backend)
	if err != nil {
		return err
	}

	http.Handle("/", http.FileServer(http.Dir(dir)))

	proxy := httputil.NewSingleHostReverseProxy(url)
	handler := func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
	http.HandleFunc("/api", handler)

	addr := fmt.Sprintf("%s:%d", host, port)
	return http.ListenAndServe(addr, nil)
}

func main() {
	if err := proxy(); err != nil {
		log.Fatal(err)
	}
}
