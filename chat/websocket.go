package chat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"

	ws "github.com/gorilla/websocket"
	"github.com/marcelbeumer/crispy-octo-goggles/chat/log"
)

var websocketHandlers = map[string]func() Event{
	"userListUpdate": func() Event { return &EventUserListUpdate{} },
	"newUser":        func() Event { return &EventNewUser{} },
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
		return fmt.Errorf("unknown event name \"%s\"", raw.Name)
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
	closed     chan struct{}
}

func (c *WebsocketConnection) SendEvent(e Event) {
	logger := c.logger
	go func() {
		m := websocketMessage{Data: e}
		eType := reflect.TypeOf(e)

		for name, handler := range websocketHandlers {
			if reflect.TypeOf(handler()) == eType {
				m.Name = name
				break
			}
		}

		if m.Name == "" {
			logger.Error(
				"can not send message type over websocket",
				map[string]any{"type": eType.String()},
			)
			return
		}

		jsonText, err := json.Marshal(&m)
		if err != nil {
			logger.Error(
				"could marshal event to json",
				map[string]any{"error": err},
			)
		}

		err = c.wsConn.WriteMessage(ws.TextMessage, jsonText)
		if err != nil {
			logger.Info(
				"websocket write error",
				map[string]any{"error": err},
			)
			return
		}
	}()
}

func (c *WebsocketConnection) ReadEvent() (Event, error) {
	select {
	case <-c.closed:
		return nil, io.EOF
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

func (c *WebsocketConnection) wsReadPump() {
	logger := c.logger
	go func() {
		for {
			if c.Closed() {
				return
			}

			// TODO look closely here and implement correct error handling
			messageType, p, err := c.wsConn.ReadMessage()
			if err != nil {
				logger.Error(
					"websocket read error",
					map[string]any{"error": err},
				)
				c.Close()
				return
			}

			logger.Debug(
				"websocket received message",
				map[string]any{"value": string(p)},
			)

			switch messageType {

			case ws.TextMessage:
				var m websocketMessage
				if err := json.Unmarshal(p, &m); err != nil {
					logger.Info(
						"could not parse message",
						map[string]any{"error": err},
					)
					continue
				}
				if m.Data == nil {
					logger.Warn("data was nil after parsing message")
					continue
				}
				logger.Info(
					"sending web socket message data",
					map[string]any{"name": m.Name},
				)
				select {
				case <-c.closed:
					return
				case c.eventOutCh <- m.Data:
					//
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
	conn.wsReadPump()
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
	logger.Info("starting server", map[string]any{"addr": addr})
	err := http.ListenAndServe(addr, http.HandlerFunc(s.handleHttp))
	return err
}

func (s *WebsocketServer) handleHttp(w http.ResponseWriter, r *http.Request) {
	logger := s.logger

	logger.Info("http request", map[string]any{"remoteAddr": r.RemoteAddr})

	username := r.URL.Query().Get("username")
	if username == "" {
		logger.Info(
			"reject connection",
			map[string]any{"reason": "no username provided"},
		)
		http.Error(w, "No username provided", http.StatusBadRequest)
		return
	}

	wsConn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Info(
			"could not upgrade connection",
			map[string]any{"error": err},
		)
		return
	}

	logger.Info(
		"new websocket connection",
		map[string]any{"remoteAddr": wsConn.RemoteAddr()},
	)

	conn := NewWebsocketConnection(wsConn, logger)
	defer conn.Close()

	if err := s.hub.ConnectUser(username, conn); err != nil {
		logger.Error(
			"could not connect to room",
			map[string]any{"remoteAddr": wsConn.RemoteAddr()},
		)
		return
	}

	for !conn.Closed() {
		//
	}

	logger.Info(
		"end of websocket connection",
		map[string]any{"remoteAddr": wsConn.RemoteAddr()},
	)
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
	logger.Info(
		"connecting to server",
		map[string]any{"serverUrl": serverUrl},
	)
	wsConn, _, err := ws.DefaultDialer.Dial(serverUrl, nil)
	if err != nil {
		return nil, err
	}
	return NewWebsocketConnection(wsConn, logger), nil
}
