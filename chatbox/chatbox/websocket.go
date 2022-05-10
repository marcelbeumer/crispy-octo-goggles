package chatbox

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sync"

	ws "github.com/gorilla/websocket"
	"github.com/marcelbeumer/crispy-octo-goggles/chatbox/log"
)

const (
	WEBSOCKET_EVENT_USER_LIST_UPDATE = "userListUpdate"
	WEBSOCKET_EVENT_NEW_USER         = "newUser"
	WEBSOCKET_EVENT_SEND_MESSAGE     = "sendMessage"
	WEBSOCKET_EVENT_NEW_MESSAGE      = "newMessage"
)

type WebsocketMessage struct {
	Name string
	Data Event
}

func (m *WebsocketMessage) UnmarshalJSON(data []byte) error {
	type JsonMessage struct {
		Name string          `json:"name"`
		Data json.RawMessage `json:"data"`
	}
	result := JsonMessage{Data: json.RawMessage{}}
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}
	m.Name = result.Name

	switch result.Name {
	case WEBSOCKET_EVENT_USER_LIST_UPDATE:
		var d EventUserListUpdate
		if err := json.Unmarshal(result.Data, &d); err != nil {
			return err
		}
		m.Data = d
	case WEBSOCKET_EVENT_NEW_USER:
		var d EventNewUser
		if err := json.Unmarshal(result.Data, &d); err != nil {
			return err
		}
		m.Data = d
	case WEBSOCKET_EVENT_SEND_MESSAGE:
		var d EventSendMessage
		if err := json.Unmarshal(result.Data, &d); err != nil {
			return err
		}
		m.Data = d
	case WEBSOCKET_EVENT_NEW_MESSAGE:
		var d EventNewMessage
		if err := json.Unmarshal(result.Data, &d); err != nil {
			return err
		}
		m.Data = d
	default:
		return fmt.Errorf("could not unmarshal websocket message with name %s", m.Name)
	}
	return nil
}

func (m *WebsocketMessage) MarshalJSON() ([]byte, error) {
	type JsonMessage struct {
		Name string `json:"name"`
		Data any    `json:"data"`
	}
	o := JsonMessage{Data: m.Data}
	switch m.Data.(type) {
	case EventUserListUpdate:
		o.Name = WEBSOCKET_EVENT_USER_LIST_UPDATE
	case EventNewUser:
		o.Name = WEBSOCKET_EVENT_NEW_USER
	case EventNewMessage:
		o.Name = WEBSOCKET_EVENT_NEW_MESSAGE
	case EventSendMessage:
		o.Name = WEBSOCKET_EVENT_SEND_MESSAGE
	default:
		return nil, fmt.Errorf("could not marshal websocket message with name %s", m.Name)
	}
	b, err := json.Marshal(o)
	return b, err
}

type WebsocketConnection struct {
	logger    log.Logger
	wsConn    *ws.Conn
	fromOther chan Event
	stop      chan struct{}
	closeOnce *sync.Once
}

// SendEvent posts event. Non-blocking, shoot and forget
func (c *WebsocketConnection) SendEvent(e Event) {
	logger := c.logger
	go func() {
		m := WebsocketMessage{Data: e}
		switch t := e.(type) {
		case EventUserListUpdate:
			m.Name = WEBSOCKET_EVENT_USER_LIST_UPDATE
		case EventNewUser:
			m.Name = WEBSOCKET_EVENT_NEW_USER
		case EventSendMessage:
			m.Name = WEBSOCKET_EVENT_NEW_MESSAGE
		case EventNewMessage:
			m.Name = WEBSOCKET_EVENT_SEND_MESSAGE
		default:
			logger.Error(
				"can not send message type over websocket",
				map[string]any{"type": reflect.TypeOf(t).String()},
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

// ReceiveEvent returns chan for Event
func (c *WebsocketConnection) ReceiveEvent() <-chan Event {
	return c.fromOther
}

// Disconnect disconnects from the server, if connected.
func (c *WebsocketConnection) Close() error {
	c.closeOnce.Do(func() {
		close(c.stop)
	})
	return c.wsConn.Close()
}

func (c *WebsocketConnection) wsReadPump() {
	logger := c.logger
	go func() {
		for {
			select {
			case <-c.stop:
				return
			default:
			}

			messageType, p, err := c.wsConn.ReadMessage()
			if err != nil {
				logger.Info(
					"websocket read error",
					map[string]any{"error": err},
				)
				return
			}
			logger.Debug(
				"websocket received message",
				map[string]any{"value": string(p)},
			)
			switch messageType {
			case ws.TextMessage:
				var m WebsocketMessage
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
				select {
				case <-c.stop:
					return
				case c.fromOther <- m.Data:
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
		logger:    logger,
		wsConn:    wsConn,
		fromOther: make(chan Event),
		stop:      make(chan struct{}),
		closeOnce: &sync.Once{},
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
	logger.Info(
		"starting server",
		map[string]any{"addr": addr},
	)
	err := http.ListenAndServe(addr, http.HandlerFunc(s.handleHttp))
	return err
}

func (s *WebsocketServer) handleHttp(w http.ResponseWriter, r *http.Request) {
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

	wsConn, err := s.upgrader.Upgrade(w, r, nil)
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

	conn := NewWebsocketConnection(wsConn, logger)

	toUser := make(chan Event)
	toHub := make(chan Event)

	// XXX maybe remove toUser/toHub API and just pass connection?

	if err := s.hub.ConnectUser(username, toUser, toHub); err != nil {
		logger.Error(
			"could not connect to room",
			map[string]any{"remoteAddr": wsConn.RemoteAddr()},
		)
		return
	}

	for {
		select {
		case e := <-toUser:
			conn.SendEvent(e)
		case e := <-conn.ReceiveEvent():
			toHub <- e
		}
	}
}

func NewWebsocketClientConnection(
	serverAddr string,
	username string,
	logger log.Logger,
) (*WebsocketConnection, error) {
	q := url.Values{}
	q.Add("username", username)
	u := url.URL{Scheme: "ws", Host: serverAddr, Path: "/", RawQuery: q.Encode()}
	serverUrl := u.String()
	fmt.Println("connecting to", serverUrl)
	wsConn, _, err := ws.DefaultDialer.Dial(serverUrl, nil)
	if err != nil {
		return nil, err
	}
	return NewWebsocketConnection(wsConn, logger), nil
}
