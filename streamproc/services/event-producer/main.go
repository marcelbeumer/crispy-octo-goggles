package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/alecthomas/kong"
)

type CLI struct {
	APIHost string `help:"API host." short:"h" default:"127.0.0.1" env:"API_HOST"`
	APIPort int    `help:"API port." short:"p" default:"9998"      env:"API_PORT"`
}

type EventBuffer struct {
	events []Event
	l      sync.RWMutex
}

func (b *EventBuffer) Append(e Event) {
	b.l.Lock()
	defer b.l.Unlock()
	b.events = append(b.events, e)
}

func (b *EventBuffer) Recover(e []Event) {
	b.l.Lock()
	defer b.l.Unlock()
	b.events = append(e, b.events...)
}

func (b *EventBuffer) Flush() []Event {
	b.l.Lock()
	defer b.l.Unlock()
	slice := b.events[:]
	b.events = make([]Event, 0)
	return slice
}

// Event is send to server for processing
type Event struct {
	// Time is timestamp in UnixMilli
	Time int64 `json:"time"`
	// Amount if number between 0-10
	Amount big.Float `json:"amount"`
}

func main() {
	cli := CLI{}
	_ = kong.Parse(
		&cli,
		kong.Name("event-producer"),
		kong.UsageOnError(),
	)

	buffer := EventBuffer{}

	go func() {
		for {
			time.Sleep(time.Second * 2)

			events := buffer.Flush()
			if len(events) == 0 {
				continue
			}

			j, err := json.Marshal(&events)
			if err != nil {
				log.Fatal(err)
			}

			// fmt.Printf("sending %s\n", string(j))

			resp, err := http.Post(
				fmt.Sprintf("http://%s:%d", cli.APIHost, cli.APIPort),
				"application/json",
				bytes.NewBuffer(j),
			)

			if err != nil {
				fmt.Printf("request error: %s\n", err)
				buffer.Recover(events)
				continue
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if resp.StatusCode >= 400 {
				fmt.Printf("request http %d: %s\n", resp.StatusCode, body)
				buffer.Recover(events)
				continue
			}
		}
	}()

	for {
		event := Event{
			Time:   time.Now().UnixMilli(),
			Amount: *big.NewFloat(rand.Float64() * 10),
		}
		buffer.Append(event)
		time.Sleep(time.Second)
	}
}
