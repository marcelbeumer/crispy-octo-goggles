# Chatbox

TODO:

- Cleanup client code
- Cleanup server code
- Fix user joining that already exists/existed
- Fix users disconnecting, should remove it from room
- Simple CLI UI for client:
  - Left panel: messages
  - Right panel: user list
  - Bottom panel: message input
- Proper structured logging, with levels
- Look at channel code again:
  - Where shoot and forget, do I really need a goroutine here
  - Where shoot and forget, add a timeout so the go routine can exit
- Unit tests
- GetData/NewMessage etc is a bit complicated, how would it be more simple and more "go"?
  - Start without generics maybe, explicit factory functions?
  - Different way of modelling messages? For ex: specific types for specific message types, interface to capture all (`type Message interface`)
