// Package chat implements a basic chat room application without protocol specifics.
//
// The main components are:
// - Events: events sent between client<->server (event.go)
// - Hub: the hub (or room) where users chat with each other (hub.go)
// - Connection: abstraction for sending events between client<->server (connection.go)
// - Frontend: (visual) interface for the end-user (gui.go, stdout.go)
//
// Protocol specifics (e.g WebSockets and gRPC) are implemented elsewhere and
// abstracted behind the Connection interface (internal/gprc, internal/websocket).
package chat
