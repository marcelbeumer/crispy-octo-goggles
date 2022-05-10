package chatbox

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/log"
)

type StdoutFrontend struct {
	logger log.Logger
	client Client
	conn   Connection
}

func (f *StdoutFrontend) Start(serverAddr string, username string) error {
	conn, err := f.client.Connect(serverAddr, username)
	if err != nil {
		return err
	}
	f.conn = conn

	defer f.conn.Close()
	stop := make(chan struct{})
	defer close(stop)

	select {
	case err = <-f.pumpNextEvent(stop):
	case err = <-f.pumpStdin(stop):
	}
	return err
}

func (f *StdoutFrontend) pumpStdin(stop <-chan struct{}) <-chan error {
	done := make(chan error)
	in := bufio.NewReader(os.Stdin)
	input := []byte{}

	go func() {
		defer close(done)

		for {
			b, err := in.ReadByte() // no way to interrupt
			if err != nil {
				continue
			}

			select {
			case <-stop:
				return
			default:
				if string(b) == "\n" {
					msg := string(input)
					input = []byte{} // reset
					f.conn.SendEvent(EventSendMessage{
						EventMeta: EventMeta{Time: time.Now()},
						Message:   msg,
					})
				} else {
					input = append(input, b)
				}
			}
		}
	}()

	return done
}

func (f *StdoutFrontend) pumpNextEvent(stop <-chan struct{}) <-chan error {
	done := make(chan error)
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			case e, ok := <-f.conn.ReceiveEvent():
				if !ok || e == nil {
					return
				}
				switch t := e.(type) {
				case EventNewUser:
					fmt.Printf(
						"[%s] <<user \"%s\" entered the room>>\n",
						t.Time.Local(),
						t.Name,
					)
				case EventNewMessage:
					fmt.Printf(
						"[%s %s] >> %s\n",
						t.Time.Local(),
						t.Sender,
						t.Message,
					)
				}
			}
		}
	}()
	return done
}

func NewStdoutClient(client Client, logger log.Logger) *StdoutFrontend {
	return &StdoutFrontend{
		logger: logger,
		client: client,
	}
}
