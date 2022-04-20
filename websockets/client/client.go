package client

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func StartClient(serverAddr string) error {
	u := url.URL{Scheme: "ws", Host: serverAddr, Path: "/"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Read error: %s", err)
				return
			}
			log.Printf("Recv: %s", msg)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {

		case <-done:
			return nil

		case t := <-ticker.C:
			err := conn.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Printf("Write error: %s", err)
				return err
			}

		case <-interrupt:
			log.Println("Interrupt")

			if err := conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
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
