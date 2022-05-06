package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	ws "github.com/gorilla/websocket"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/channels"
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

	defer wsConn.Close()
	log.Printf("new websocket conn with %s\n", wsConn.RemoteAddr())

	ch := channels.NewChannels()
	if err := s.room.Connect(username, channels.NewChannelsForRoom(ch)); err != nil {
		log.Printf("could not connect to room: %s\n", err)
		return
	}

	log.Printf("user %s connected from %s\n", username, wsConn.RemoteAddr())

	done := make(chan struct{})
	defer close(done)
	wsDone := s.wsReadPump(wsConn, ch, done)

	for {
		select {
		case <-wsDone:
			if err := s.room.Disconnect(username); err != nil {
				panic(err)
			}
			return
		case m := <-ch.ToUser:
			jsonText, err := json.Marshal(m)
			if err != nil {
				panic(err)
			}

			err = wsConn.WriteMessage(ws.TextMessage, jsonText)
			if err != nil {
				log.Printf("Write error: %s", err)
				return
			}
		}
	}
}

func (s *Server) wsReadPump(
	wsConn *ws.Conn,
	ch *channels.Channels,
	stop chan struct{},
) (done chan error) {
	done = make(chan error)
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
				select {
				case <-done:
					return
				case ch.ToRoom <- m:
				}
			default:
				log.Printf("websocket ignoring message type: %d\n", messageType)
			}
		}
	}()
	return
}

func NewServer() *Server {
	return &Server{
		room: room.NewRoom(),
	}
}
