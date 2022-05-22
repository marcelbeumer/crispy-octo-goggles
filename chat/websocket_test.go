package chat

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/marcelbeumer/crispy-octo-goggles/chat/log"
	"github.com/marcelbeumer/crispy-octo-goggles/chat/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebsocketServerHandleHttp(t *testing.T) {

	t.Run("GET / should 400 not having an username", func(t *testing.T) {
		wsServer := NewWebsocketServer(&log.NoopLoggerAdapter{})
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
		wsServer := NewWebsocketServer(test.NewTestLogger(true))
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

	t.Run("GET /?username=User should upgrade the connection", func(t *testing.T) {
		wsServer := NewWebsocketServer(test.NewTestLogger(true))
		server := httptest.NewServer(http.HandlerFunc(wsServer.handleHttp))

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/?username=User"
		ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if ws != nil {
			defer ws.Close()
		}

		require.NoError(t, err)
	})
}
