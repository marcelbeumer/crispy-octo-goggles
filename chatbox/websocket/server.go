package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"

	ws "github.com/gorilla/websocket"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/channels"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/log"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/message"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/room"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Server struct {
	logger log.Logger
	room   *room.Room
}

func (s *Server) Start(addr string) error {
	fmt.Println("starting server on", addr)
	err := http.ListenAndServe(addr, http.HandlerFunc(s.handleHttp))
	return err
}

func (s *Server) handleHttp(w http.ResponseWriter, r *http.Request) {
	logger := s.logger

	logger.Info(
		"http request",
		map[string]any{"remoteAddr": r.RemoteAddr},
	)

	username := r.URL.Query().Get("username")
	if username == "" {
		logger.Info(
			"reject connection",
			map[string]any{"reason": "no username provided"},
		)
		http.Error(w, "No username provided", http.StatusBadRequest)
		return
	}

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Info(
			"could not upgrade connection",
			map[string]any{"error": err},
		)
		return
	}

	defer wsConn.Close()
	logger.Info(
		"new websocket connection",
		map[string]any{"remoteAddr": wsConn.RemoteAddr()},
	)

	ch := channels.NewChannels()
	if err := s.room.Connect(username, channels.NewChannelsForRoom(ch)); err != nil {
		logger.Error(
			"could not connect to room",
			map[string]any{"remoteAddr": wsConn.RemoteAddr()},
		)
		return
	}

	logger.Info(
		"user connected",
		map[string]any{
			"username":   username,
			"remoteAddr": wsConn.RemoteAddr(),
		},
	)

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
				logger.Info(
					"websocket write error",
					map[string]any{"error": err},
				)
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
	logger := s.logger
	done = make(chan error)
	go func() {
		defer close(done)
		for {
			messageType, p, err := wsConn.ReadMessage()
			if err != nil {
				logger.Info(
					"websocket read error",
					map[string]any{"error": err},
				)
				return
			}
			switch messageType {
			case ws.TextMessage:
				logger.Debug(
					"websocket received message",
					map[string]any{"value": string(p)},
				)

				m := message.Message{}
				if err := json.Unmarshal(p, &m); err != nil {
					logger.Info(
						"could not parse message",
						map[string]any{"error": err},
					)
					continue
				}
				select {
				case <-done:
					return
				case ch.ToRoom <- m:
				}
			default:
				logger.Info(
					"websocket ignoring message type: %d",
					map[string]any{"messageType": messageType},
				)
			}
		}
	}()
	return
}

func NewServer(logger log.Logger) *Server {
	return &Server{
		logger: logger,
		room:   room.NewRoom(),
	}
}
