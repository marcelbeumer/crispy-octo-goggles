package websocket

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/marcelbeumer/go-playground/gochat/internal/log"
	"github.com/marcelbeumer/go-playground/gochat/internal/util/now"
	"github.com/marcelbeumer/go-playground/gochat/internal/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {

	t.Run("GET / should 400 not having an username", func(t *testing.T) {
		wsServer := NewServer(&log.NoopLoggerAdapter{})
		req := httptest.NewRequest("get", "/", nil)
		w := httptest.NewRecorder()
		wsServer.handleHttp(w, req)
		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
		assert.Equal(t, "No username provided\n", string(body))
	})

	t.Run("websocket on / should 400 not having an username", func(t *testing.T) {
		wsServer := NewServer(test.NewTestLogger(true))
		server := httptest.NewServer(http.HandlerFunc(wsServer.handleHttp))

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/"
		ws, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if ws != nil {
			defer ws.Close()
		}

		require.Error(t, err)
		assert.Equal(t, "websocket: bad handshake", err.Error())

		require.NotNil(t, resp)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
		assert.Equal(t, "No username provided\n", string(body))
	})

	t.Run("GET /?username=User should connect a websocket", func(t *testing.T) {
		wsServer := NewServer(test.NewTestLogger(true))
		server := httptest.NewServer(http.HandlerFunc(wsServer.handleHttp))

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/?username=User"
		wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer wsConn.Close()
	})

	t.Run("sets up hub communication", func(t *testing.T) {
		nowStub := now.SetupStub()
		nowStub.Frozen = true
		t.Cleanup(func() {
			now.ClearStub()
		})

		wsServer := NewServer(test.NewTestLogger(true))

		server := httptest.NewServer(http.HandlerFunc(wsServer.handleHttp))
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/?username=User"

		nowStub.Inc()
		wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer wsConn.Close()

		wsConn.SetReadDeadline(time.Now().Add(time.Second))
		messageType, p, err := wsConn.ReadMessage()
		require.NoError(t, err)
		require.Equal(t, websocket.TextMessage, messageType)
		msg := `{"name":"connected","data":{"time":"1970-01-01T01:00:01+01:00","users":["User"]}}`
		require.Equal(t, msg, string(p))

		nowStub.Inc()
		msg = `{"name":"sendMessage","data":{"time":"1970-01-01T01:00:02+01:00","message":"Hello"}}`
		wsConn.SetWriteDeadline(time.Now().Add(time.Second))
		err = wsConn.WriteMessage(websocket.TextMessage, []byte(msg))
		require.NoError(t, err)

		nowStub.Inc()
		wsConn.SetReadDeadline(time.Now().Add(time.Second))
		messageType, p, err = wsConn.ReadMessage()
		require.NoError(t, err)
		require.Equal(t, websocket.TextMessage, messageType)
		msg = `{"name":"newMessage","data":{"time":"1970-01-01T01:00:03+01:00","sender":"User","message":"Hello"}}`
		require.Equal(t, msg, string(p))
	})
}
