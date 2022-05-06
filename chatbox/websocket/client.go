package websocket

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	ws "github.com/gorilla/websocket"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/channels"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/message"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/user"
)

type Client struct {
	Username   string
	ServerAddr string
	wsConn     *ws.Conn
	channels   *channels.Channels
	user       *user.User
	stdinCh    chan byte
	disconnect chan bool
}

func (c *Client) Connect(serverAddr string, username string) error {
	c.Username = username
	c.ServerAddr = serverAddr

	if err := c.connectWs(); err != nil {
		return err
	}
	defer c.Disconnect()

	c.user = user.NewUser()
	c.channels = channels.NewChannels()
	c.user.Connect(channels.NewChannelsForUser(c.channels))
	c.stdinCh = make(chan byte)
	c.disconnect = make(chan bool)

	stop := make(chan struct{})
	var err error

	select {
	case <-c.disconnect:
	case err = <-c.wsReadPump(stop):
	case err = <-c.wsWritePump(stop):
	case err = <-c.stdinPump(stop):
	case err = <-c.waitInterrupt(stop):
	}

	close(stop)
	return err
}

func (c *Client) Disconnect() error {
	var err error
	if c.disconnect != nil {
		c.disconnect <- true
	}
	if c.wsConn != nil {
		err = c.wsConn.Close()
		c.wsConn = nil
	}
	return err
}

func (c *Client) stdinPump(stop <-chan struct{}) (done chan error) {
	done = make(chan error)
	in := bufio.NewReader(os.Stdin)
	ch := c.stdinCh

	go func() {
		defer close(done)
		for {
			// FIXME could wait here forever but no idea
			// how to cancel/abort a Reader from stdin
			b, err := in.ReadByte()
			if err != nil {
				continue
			}
			select {
			case <-stop:
				return
			case ch <- b:
				//
			}
		}
	}()

	return
}

func (c *Client) wsReadPump(stop <-chan struct{}) (done chan error) {
	done = make(chan error)
	toUser := c.channels.ToUser
	wsConn := c.wsConn

	go func() {
		defer close(done)
		for {
			messageType, p, err := wsConn.ReadMessage()
			if err != nil {
				log.Printf("websocket read error: %s\n", err)
				return
			}
			switch messageType {
			case ws.TextMessage:
				// log.Printf("websocket received message: %s\n", string(p))
				m := message.Message{}
				if err := json.Unmarshal(p, &m); err != nil {
					log.Printf("could not parse message: %s\n", err)
				}
				select {
				case <-stop:
					return
				case toUser <- m:
					//
				}
			default:
				log.Printf("websocket ignoring message type: %d\n", messageType)
			}
		}
	}()

	return
}

func (c *Client) wsWritePump(stop <-chan struct{}) (done chan error) {
	done = make(chan error)
	input := []byte{}
	stdinCh := c.stdinCh
	wsConn := c.wsConn

	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			case s := <-stdinCh:
				if string(s) == "\n" {
					msg := string(input)
					input = []byte{} // reset
					m, err := message.NewMessage(message.SEND_MESSAGE, msg)
					if err != nil {
						panic(err)
					}
					jsonText, err := json.Marshal(m)
					if err != nil {
						panic(err)
					}
					// log.Printf("websocket sending message: %s\n", jsonText)
					err = wsConn.WriteMessage(ws.TextMessage, jsonText)
					if err != nil {
						log.Printf("write error: %s", err)
						return
					}
				} else {
					input = append(input, s)
				}
			}
		}
	}()

	return
}

func (c *Client) waitInterrupt(stop <-chan struct{}) (done chan error) {
	done = make(chan error)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	wsConn := c.wsConn

	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			case <-interrupt:
				log.Println("interrupt")

				if err := wsConn.WriteMessage(
					ws.CloseMessage,
					ws.FormatCloseMessage(ws.CloseNormalClosure, ""),
				); err != nil {
					return
				}

				select {
				case <-stop:
				case <-time.After(time.Second):
				}
			}
		}
	}()
	return
}

func (c *Client) connectWs() error {
	if c.wsConn != nil {
		return errors.New("already connected")
	}
	q := url.Values{}
	q.Add("username", c.Username)
	u := url.URL{Scheme: "ws", Host: c.ServerAddr, Path: "/", RawQuery: q.Encode()}
	serverUrl := u.String()
	fmt.Println("connecting to", serverUrl)
	wsConn, _, err := ws.DefaultDialer.Dial(serverUrl, nil)
	if err != nil {
		return err
	}
	c.wsConn = wsConn
	return nil
}

func NewClient() *Client {
	return &Client{}
}
