package chat

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/marcelbeumer/crispy-octo-goggles/chat/log"
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
	case err = <-fnCh(func() error { return f.pumpEvents(stop) }):
	case err = <-fnCh(func() error { return f.pumpStdin(stop) }):
	}
	return err
}

func (f *StdoutFrontend) pumpStdin(stop <-chan struct{}) error {
	in := bufio.NewReader(os.Stdin)
	input := []byte{}
	for {
		select {
		case <-stop:
			return nil
		default:
			b, err := in.ReadByte() // no way to interrupt
			if err != nil {
				return err
			}
			if string(b) == "\n" {
				msg := string(input)
				input = []byte{} // reset
				err := f.conn.SendEvent(&EventSendMessage{
					EventMeta: EventMeta{Time: time.Now()},
					Message:   msg,
				})
				if err != nil {
					return err
				}
			} else {
				input = append(input, b)
			}
		}
	}
}

func (f *StdoutFrontend) pumpEvents(stop <-chan struct{}) error {
	for {
		select {
		case <-stop:
			return nil
		default:
			logger := f.logger
			e, err := f.conn.ReadEvent()
			if err != nil {
				return err
			}

			switch t := e.(type) {
			case *EventConnected:
				//
			case *EventUserListUpdate:
				//
			case *EventUserEnter:
				fmt.Printf(
					"[%s] <<user \"%s\" entered the room>>\n",
					t.Time.Local(),
					t.Name,
				)
			case *EventUserLeave:
				fmt.Printf(
					"[%s] <<user \"%s\" left the room>>\n",
					t.Time.Local(),
					t.Name,
				)
			case *EventNewMessage:
				fmt.Printf(
					"[%s %s] >> %s\n",
					t.Time.Local(),
					t.Sender,
					t.Message,
				)
			default:
				logger.Warnw(
					"unhandled event type",
					"type", reflect.TypeOf(e).String())
			}
		}
	}
}

func NewStdoutFrontend(conn Connection, logger log.Logger) *StdoutFrontend {
	return &StdoutFrontend{
		logger: logger,
		conn:   conn,
	}
}
