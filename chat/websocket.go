package chat

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sync"

	ws "github.com/gorilla/websocket"
	"github.com/marcelbeumer/crispy-octo-goggles/chat/log"
)

var websocketHandlers = map[string]func() Event{
	"userListUpdate": func() Event { return &EventUserListUpdate{} },
	"userEnter":      func() Event { return &EventUserEnter{} },
	"userLeave":      func() Event { return &EventUserLeave{} },
	"sendMessage":    func() Event { return &EventSendMessage{} },
	"newMessage":     func() Event { return &EventNewMessage{} },
}

type websocketMessage struct {
	Name string `json:"name"`
	Data Event  `json:"data"`
}

type websocketMessageRaw struct {
	Name string           `json:"name"`
	Data *json.RawMessage `json:"data"`
}

func (m *websocketMessage) UnmarshalJSON(data []byte) error {
	raw := websocketMessageRaw{Data: &json.RawMessage{}}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	handler := websocketHandlers[raw.Name]
	if handler == nil {
		return fmt.Errorf(`unknown event name "%s"`, raw.Name)
	}

	m.Name = raw.Name
	m.Data = handler()

	if err := json.Unmarshal(*raw.Data, &m.Data); err != nil {
		return err
	}

	return nil
}

type WebsocketConnection struct {
	logger     log.Logger
	wsConn     *ws.Conn
	eventOutCh chan Event
	l          sync.RWMutex
	closed     chan struct{}
}

func (c *WebsocketConnection) SendEvent(e Event) error {
	c.l.Lock()
	defer c.l.Unlock()

	select {
	case <-c.closed:
		return ErrConnectionClosed
	default:
	}

	m := websocketMessage{Data: e}
	eType := reflect.TypeOf(e)
	eTypeStr := eType.String()

	for name, handler := range websocketHandlers {
		if reflect.TypeOf(handler()) == eType {
			m.Name = name
			break
		}
	}

	if m.Name == "" {
		return fmt.Errorf("unknown event type <%s>", eTypeStr)
	}

	jsonText, err := json.Marshal(&m)
	if err != nil {
		return fmt.Errorf(
			"could not marshal event with type <%s>: %w",
			eTypeStr, err)
	}

	err = c.wsConn.WriteMessage(ws.TextMessage, jsonText)

	if err != nil {
		return fmt.Errorf(
			"error writing event to ws with type <%s>: %w",
			eTypeStr, err)
	}

	return nil
}

func (c *WebsocketConnection) ReadEvent() (Event, error) {
	select {
	case <-c.closed:
		return nil, ErrConnectionClosed
	case e := <-c.eventOutCh:
		return e, nil
	}
}

func (c *WebsocketConnection) Closed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

func (c *WebsocketConnection) Close() error {
	select {
	case <-c.closed:
	default:
		close(c.closed)
		return c.wsConn.Close()
	}
	return nil
}

func (c *WebsocketConnection) wsReadPump() error {
	for {
		messageType, p, err := c.wsConn.ReadMessage()
		if err != nil {
			return err
		}

		switch messageType {
		case ws.TextMessage:
			var m websocketMessage
			if err := json.Unmarshal(p, &m); err != nil {
				return fmt.Errorf(`could not unmarshal message: %w`, err)
			}
			if m.Data == nil {
				return fmt.Errorf("data was nil after parsing message")
			}
			select {
			case <-c.closed:
				return ErrConnectionClosed
			case c.eventOutCh <- m.Data:
				//
			}
		}
	}
}

func NewWebsocketConnection(
	wsConn *ws.Conn,
	logger log.Logger,
) *WebsocketConnection {
	conn := WebsocketConnection{
		logger:     logger,
		wsConn:     wsConn,
		eventOutCh: make(chan Event),
		closed:     make(chan struct{}),
	}
	go func() {
		defer conn.Close()
		err := conn.wsReadPump()
		if errors.Is(err, ErrConnectionClosed) {
			logger.Infow("websocket pump closed")
		} else {
			logger.Errorw("websocket pump error", log.Error(err))
		}
	}()
	return &conn
}

type WebsocketServer struct {
	logger   log.Logger
	upgrader ws.Upgrader
	hub      *Hub
}

func NewWebsocketServer(logger log.Logger) *WebsocketServer {
	return &WebsocketServer{
		logger: logger,
		hub:    NewHub(logger),
		upgrader: ws.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (s *WebsocketServer) Start(addr string) error {
	logger := s.logger
	logger.Infow("starting server", "addr", addr)
	err := http.ListenAndServe(addr, http.HandlerFunc(s.handleHttp))
	return err
}

func (s *WebsocketServer) handleHttp(w http.ResponseWriter, r *http.Request) {
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
	conn := NewWebsocketConnection(wsConn, logger)
	defer conn.Close()

	if err := s.hub.ConnectUser(username, conn); err != nil {
		logger.Errorw(
			"user disconnected with error",
			log.Error(err),
		)
	}

	logger.Infow("end of websocket connection")
}

func NewWebsocketClientConnection(
	serverAddr string,
	username string,
	logger log.Logger,
) (*WebsocketConnection, error) {
	q := url.Values{"username": []string{username}}
	u := url.URL{
		Scheme:   "ws",
		Host:     serverAddr,
		Path:     "/",
		RawQuery: q.Encode(),
	}
	serverUrl := u.String()
	logger.Infow(
		"connecting to server",
		"serverUrl", serverUrl)
	wsConn, _, err := ws.DefaultDialer.Dial(serverUrl, nil)
	if err != nil {
		return nil, err
	}
	return NewWebsocketConnection(wsConn, logger), nil
}
