package websocket

import (
	"net/url"

	ws "github.com/gorilla/websocket"
	"github.com/marcelbeumer/go-playground/gochat/internal/log"
)

func NewClientConnection(
	serverAddr string,
	username string,
	logger log.Logger,
) (*Connection, error) {
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
	return NewConnection(wsConn, logger), nil
}
