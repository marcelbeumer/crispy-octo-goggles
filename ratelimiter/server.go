package main

import (
	"fmt"
	"net/http"
)

func StartServer(addr string) error {
	fmt.Printf("Starting server on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		return err
	}
	return nil
}
