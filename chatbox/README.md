# Chatbox

TODO:

- Simple CLI UI for client:
  - Left panel: messages
  - Right panel: user list
  - Bottom panel: message input
- Unit tests
- GetData/NewMessage etc is a bit complicated, how would it be more simple and more "go"?
  - Start without generics maybe, explicit factory functions?
  - Different way of modelling messages? For ex: specific types for specific message types, interface to capture all (`type Message interface`)

Refactor for UI

- Move main to cmd/main.go
- Core packages can move to toplevel
- Some stay in other package like websocket (later grpc), ui
- Replace messages with Events: `type Event interface` + specific `type Event<Thing> struct` just like gocui tcell
  - Makes type switches easy
  - Much more "Go"
  - Transfer protocol, websocket/grpc/... can determine how to transfer... in that domain
- `channels.go` is fluff, I think it can completely go away; just pass channels directly to whoever needs it (? maybe overlooking smth)
- At the moment a `User` does not implement real business logic yet, that's why User & UI feel redundant
  - Principle: do not introduce abstractions that are not needed
  - So you can make a `StandardClient` and a `GUIClient`, they will both do the same but render differently

Probably start from below
