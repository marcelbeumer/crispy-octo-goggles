# Chatbox

TODO:

- Simple CLI UI for client:
  - Left panel: messages
  - Right panel: user list
  - Bottom panel: message input
- Make client go routines exit properly by using read deadlines on stdin (if that works)
- Unit tests
- GetData/NewMessage etc is a bit complicated, how would it be more simple and more "go"?
  - Start without generics maybe, explicit factory functions?
  - Different way of modelling messages? For ex: specific types for specific message types, interface to capture all (`type Message interface`)
