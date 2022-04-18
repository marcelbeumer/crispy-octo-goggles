package client

import (
	"bufio"
	"fmt"
	"os"
)

type Client struct {
	doneCh chan struct{}
}

func (s *Client) Connect(serverAddr string) error {
	// r := s.initRouting()
	// err := http.ListenAndServe(addr, r)
	return nil
}

func (s *Client) Start() chan struct{} {
	s.doneCh = make(chan struct{})

	stdinChan := make(chan string)

	go (func() {
		in := bufio.NewReader(os.Stdin)
		for {
			if s.doneCh == nil {
				return
			}
			b, err := in.ReadByte()
			if err != nil {
				continue
			}
			stdinChan <- string(b)
		}
	})()

	go (func() {
		input := []string{}
		for {
			select {
			case s := <-stdinChan:
				if s == "\n" {
					msg := input
					input = []string{} // reset
					fmt.Printf("Will send: %s\n", msg)
				} else {
					input = append(input, s)
				}
			case <-s.doneCh:
				s.doneCh = nil
				return
			}
		}
	})()

	return s.doneCh
}

func NewClient() *Client {
	s := Client{}
	return &s
}
