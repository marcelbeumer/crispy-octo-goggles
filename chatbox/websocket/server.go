package websocket

import (
	"fmt"
	"log"
	"net/http"

	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func StartServer(addr string) error {
	fmt.Println("Starting server on", addr)
	err := http.ListenAndServe(addr, http.HandlerFunc(handleHttp))
	if err != nil {
		return err
	}
	return nil
}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request from %s\n", r.RemoteAddr)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Could not upgrade connection: %s\n", err)
		return
	}
	log.Printf("New websocket conn with %s\n", conn.RemoteAddr())

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %s\n", err)
			return
		}
		log.Printf("Forwarding message of type %d: %s\n", messageType, p)
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Printf("Write error: %s\n", err)
			return
		}
	}
}
