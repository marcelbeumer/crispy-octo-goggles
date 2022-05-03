package websocket

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	ws "github.com/gorilla/websocket"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/connection"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/message"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/user"
)

func getServerUrl(serverAddr string, username string) string {
	q := url.Values{}
	q.Add("username", username)
	u := url.URL{Scheme: "ws", Host: serverAddr, Path: "/", RawQuery: q.Encode()}
	return u.String()
}

func StartClient(serverAddr string, username string) error {
	serverUrl := getServerUrl(serverAddr, username)
	fmt.Println("connecting to", serverUrl)
	wsConn, _, err := ws.DefaultDialer.Dial(serverUrl, nil)
	if err != nil {
		return err
	}
	defer wsConn.Close()

	u := user.NewUser()
	channels := connection.NewChannels()
	u.ConnectRoom(connection.NewConnectionForUser(channels))
	done := make(chan struct{})

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
				go func() {
					channels.ToUser <- m
				}()
			default:
				log.Printf("websocket ignoring message type: %d\n", messageType)
			}
		}
	}()

	stdinChan := make(chan byte)

	go (func() {
		in := bufio.NewReader(os.Stdin)
		for {
			if done == nil {
				return
			}
			b, err := in.ReadByte()
			if err != nil {
				continue
			}
			stdinChan <- b
		}
	})()

	go (func() {
		input := []byte{}
		for {
			select {
			case s := <-stdinChan:
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
			case <-done:
				return
			}
		}
	})()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		select {

		case <-done:
			return nil

		case <-interrupt:
			log.Println("interrupt")

			if err := wsConn.WriteMessage(
				ws.CloseMessage,
				ws.FormatCloseMessage(ws.CloseNormalClosure, ""),
			); err != nil {
				return err
			}

			select {
			case <-done:
			case <-time.After(time.Second):
			}

			return nil
		}
	}
}
