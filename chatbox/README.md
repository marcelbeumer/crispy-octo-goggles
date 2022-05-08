# Chatbox

TODO:

- Control log level from cli (maybe just `-v` for info and `-vv` for debug)
- Simple CLI UI for client:
  - Left panel: messages
  - Right panel: user list
  - Bottom panel: message input
- Unit tests
- GetData/NewMessage etc is a bit complicated, how would it be more simple and more "go"?
  - Start without generics maybe, explicit factory functions?
  - Different way of modelling messages? For ex: specific types for specific message types, interface to capture all (`type Message interface`)
