package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	ws "github.com/gorilla/websocket"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/connection"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/message"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/room"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Server struct {
	room *room.Room
}

func (s *Server) Start(addr string) error {
	fmt.Println("starting server on", addr)
	err := http.ListenAndServe(addr, http.HandlerFunc(s.handleHttp))
	return err
}

func (s *Server) handleHttp(w http.ResponseWriter, r *http.Request) {
	log.Printf("request from %s\n", r.RemoteAddr)
	username := r.URL.Query().Get("username")
	if username == "" {
		log.Printf("no username provided")
		http.Error(w, "No username provided", http.StatusBadRequest)
		return
	}
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("could not upgrade connection: %s\n", err)
		return
	}
	log.Printf("new websocket conn with %s\n", wsConn.RemoteAddr())

	channels := connection.NewChannels()
	s.room.ConnectUser(username, connection.NewConnectionForRoom(channels))
	done := make(chan struct{})

	log.Printf("user %s connected from %s\n", username, wsConn.RemoteAddr())

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
				log.Printf("websocket received message: %s\n", string(p))
				m := message.Message{}
				if err := json.Unmarshal(p, &m); err != nil {
					log.Printf("could not parse message: %s\n", err)
				}
				go func() {
					channels.ToRoom <- m
				}()
			default:
				log.Printf("websocket ignoring message type: %d\n", messageType)
			}
		}
	}()

	for {
		select {

		case m := <-channels.ToUser:
			jsonText, err := json.Marshal(m)
			if err != nil {
				panic(err)
			}

			err = wsConn.WriteMessage(ws.TextMessage, jsonText)
			if err != nil {
				log.Printf("Write error: %s", err)
				return
			}

		case <-done:
			return
		}
	}
}

func NewServer() *Server {
	return &Server{
		room: room.NewRoom(),
	}
}

func StartServer(addr string) (*Server, error) {
	s := NewServer()
	return s, s.Start(addr)
}
