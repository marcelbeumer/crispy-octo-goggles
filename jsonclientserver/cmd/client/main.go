package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/marcelbeumer/go-playground/jsonclientserver/internal/client"
)

func exit(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}

func main() {
	var timeoutSec int
	var requestTotal int
	var batchSize int
	var serverAddr string

	flag.StringVar(&serverAddr, "server", "http://localhost:8080/json", "server address")
	flag.IntVar(&timeoutSec, "timeout", 60, "timeout after <timeout> seconds")
	flag.IntVar(&requestTotal, "requests", 100, "total request")
	flag.IntVar(&batchSize, "batch", 25, "request batch size")
	flag.Parse()

	cl := client.NewClient(client.ClientOpts{
		ServerAddr:   serverAddr,
		RequestTotal: requestTotal,
		BatchSize:    batchSize,
	})

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(timeoutSec)*time.Second)

	defer cancel()

	if err := cl.Run(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
