package websocket

import (
	"github.com/marcelbeumer/go-playground/gochat/internal/chat"
)

var handlers = map[string]func() chat.Event{
	"connected":      func() chat.Event { return &chat.EventConnected{} },
	"userListUpdate": func() chat.Event { return &chat.EventUserListUpdate{} },
	"userEnter":      func() chat.Event { return &chat.EventUserEnter{} },
	"userLeave":      func() chat.Event { return &chat.EventUserLeave{} },
	"sendMessage":    func() chat.Event { return &chat.EventSendMessage{} },
	"newMessage":     func() chat.Event { return &chat.EventNewMessage{} },
}
