package chat

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"reflect"
	"time"

	"github.com/marcelbeumer/crispy-octo-goggles/chat/logging"
)

type StdoutFrontend struct {
	logger logging.Logger
	conn   Connection
}

func (f *StdoutFrontend) Start() error {
	stop := make(chan struct{})
	defer close(stop)

	var err error
	select {
	case <-f.pumpEvents(stop):
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
					f.conn.SendEvent(&EventSendMessage{
						EventMeta: EventMeta{time: time.Now()},
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

func (f *StdoutFrontend) pumpEvents(stop <-chan struct{}) <-chan struct{} {
	done := make(chan struct{})
	logger := f.logger
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			default:
			}

			e, err := f.conn.ReadEvent()

			if err == io.EOF {
				logger.Info("read event EOF")
				return
			}

			if err != nil {
				logger.Error("read event error (sleep)",
					map[string]any{
						"error": err.Error(),
					})
				time.Sleep(time.Second)
				continue
			}

			switch t := e.(type) {
			case *EventUserListUpdate:
				//
			case *EventNewUser:
				fmt.Printf(
					"[%s] <<user \"%s\" entered the room>>\n",
					t.time.Local(),
					t.Name,
				)
			case *EventNewMessage:
				fmt.Printf(
					"[%s %s] >> %s\n",
					t.time.Local(),
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
	}()
	return done
}

func NewStdoutFrontend(conn Connection, logger logging.Logger) *StdoutFrontend {
	return &StdoutFrontend{
		logger: logger,
		conn:   conn,
	}
}
