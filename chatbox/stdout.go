package chatbox

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/log"
)

type StdoutFrontend struct {
	logger log.Logger
	conn   Connection
}

func (f *StdoutFrontend) Start() error {
	stop := make(chan struct{})
	defer close(stop)

	var err error
	select {
	case err = <-f.pumpEvents(stop):
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

func (f *StdoutFrontend) pumpEvents(stop <-chan struct{}) <-chan error {
	logger := f.logger
	done := make(chan error)
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			case <-f.conn.Closed():
				return
			case e := <-f.conn.ReceiveEvent():
				switch t := e.(type) {
				case EventUserListUpdate:
					//
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
				default:
					logger.Warn("unhandled event type",
						map[string]any{
							"type": reflect.TypeOf(e).String(),
						})
				}
			}
		}
	}()
	return done
}

func NewStdoutFrontend(conn Connection, logger log.Logger) *StdoutFrontend {
	return &StdoutFrontend{
		logger: logger,
		conn:   conn,
	}
}
