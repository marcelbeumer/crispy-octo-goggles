package websocket

import (
	"net/http"

	ws "github.com/gorilla/websocket"
	"github.com/marcelbeumer/crispy-octo-goggles/chat"
	"github.com/marcelbeumer/crispy-octo-goggles/chat/log"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Server struct {
	logger   log.Logger
	upgrader ws.Upgrader
	hub      *chat.Hub
}

func NewServer(logger log.Logger) *Server {
	return &Server{
		logger:   logger,
		hub:      chat.NewHub(logger),
		upgrader: upgrader,
	}
}

func (s *Server) Start(addr string) error {
	logger := s.logger
	logger.Infow("starting server", "addr", addr)
	err := http.ListenAndServe(addr, http.HandlerFunc(s.handleHttp))
	return err
}

func (s *Server) handleHttp(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.With("remoteAddr", r.RemoteAddr)
	logger.Info("http request")

	username := r.URL.Query().Get("username")
	if username == "" {
		logger.Infow(
			"reject connection",
			"reason", "no username provided",
		)
		http.Error(w, "No username provided", http.StatusBadRequest)
		return
	}

	wsConn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Infow(
			"could not upgrade connection",
			log.Error(err),
		)
		return
	}

	logger.Infow("new websocket connection")
	conn := NewConnection(wsConn, logger)
	userId, err := s.hub.Connect(username, conn)
	if err != nil {
		logger.Errorw(
			"could not connect to hub",
			log.Error(err),
		)
	}

	defer conn.Close(nil)
	defer s.hub.Disconnect(userId)

	if err := conn.Wait(); err != nil {
		logger.Errorw(
			"user disconnected with error",
			log.Error(err),
		)
	}

	logger.Infow("end of websocket connection")
}
