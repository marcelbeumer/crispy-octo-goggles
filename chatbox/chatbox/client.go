package chatbox

type Client interface {
	// Connect connects to the server. Blocks until connected.
	Connect(serverAddr string, username string) (Connection, error)
}

type Connection interface {
	// SendEvent posts event. Non-blocking, shoot and forget
	SendEvent(e Event)
	// ReceiveEvent returns chan for Event
	ReceiveEvent() <-chan Event
	// Disconnect disconnects from the server, if connected.
	// Blocks until disconnected.
	Close() error
}
