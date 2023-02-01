package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"sync"
	"time"
)

type ClientOpts struct {
	ServerAddr   string
	RequestTotal int
	BatchSize    int
}

type Client struct {
	Opts ClientOpts
}

func (c *Client) doRequest(ctx context.Context, jsonData []byte) error {
	if c.Opts.ServerAddr == "" {
		return errors.New("please set server address")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Opts.ServerAddr, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	res, err := client.Do(req)

	resJson, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var resObj any
	err = json.Unmarshal(resJson, &resObj)
	if err != nil {
		return err
	}

	reformatted, err := json.Marshal(resObj)
	if err != nil {
		return err
	}

	if string(reformatted) != string(jsonData) {
		return errors.New("JSON from response is not the same we sent.")
	}

	return nil
}

func (c *Client) Run(ctx context.Context) error {
	var requestTotal = c.Opts.RequestTotal
	var batchSize = c.Opts.BatchSize
	var mu sync.Mutex
	var requestCount int
	var errs []error

	if requestTotal == 0 {
		return errors.New("total requests must be >=1")
	}

	if batchSize == 0 {
		return errors.New("batch size must be >=1")
	}

	updateStatus := func() {
		fmt.Print("\033[2K\r") // Clear line.
		fmt.Printf("Request %d/%d (errors: %d)", requestCount, requestTotal, len(errs))
	}

	updateStatus()

	jsonObj := map[string]string{"hello": "world"}
	jsonData, err := json.Marshal(jsonObj)
	if err != nil {
		return err
	}

	for requestCount < requestTotal {
		var wg sync.WaitGroup
		next := int(math.Max(0,
			math.Min(
				float64(batchSize),
				float64(requestTotal-requestCount))))

		done := func(err error) {
			mu.Lock()
			defer mu.Unlock()
			wg.Done()
			requestCount++
			if err != nil {
				errs = append(errs, err)
			}
			updateStatus()
		}

		for i := 0; i < next; i++ {
			wg.Add(1)
			go func() {
				done(c.doRequest(ctx, jsonData))
			}()
		}

		wg.Wait()
	}

	fmt.Print("\n")

	if len(errs) > 0 {
		fmt.Println("Errors:")
		for i, err := range errs {
			fmt.Printf("request #%d: %s\n", i+1, err.Error())
		}
	}

	return nil
}

func NewClient(opts ClientOpts) *Client {
	return &Client{Opts: opts}
}
