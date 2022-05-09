package ui

import (
	"time"

	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/log"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/message"
)

type TestClient struct {
	logger  log.Logger
	msgChan chan message.Message
	stop    chan bool
}

func (t *TestClient) Start() {
	go func() {
		for {
			select {
			case <-t.stop:
				return
			case <-time.After(time.Second * 2):
				msg, err := message.NewMessage(message.NEW_MESSAGE, message.NewMessageData{
					Sender:  "Ghost",
					Message: "Hello there",
					Time:    time.Now(),
				})
				if err != nil {
					t.logger.Error(err.Error())
				} else {
					t.msgChan <- msg
				}
			}
		}
	}()
}

func (t *TestClient) SendMessage(m message.Message) {
	time.Sleep(time.Millisecond * 100)
	go func() {
		t.msgChan <- m
	}()
}

func (t *TestClient) ReceiveMessage() <-chan message.Message {
	return t.msgChan
}

func (t *TestClient) Stopped() <-chan struct{} {
	return make(chan struct{})
}

func NewTestClient(logger log.Logger) *TestClient {
	return &TestClient{
		logger:  logger,
		msgChan: make(chan message.Message),
		stop:    make(chan bool),
	}
}
