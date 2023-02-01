# jsonclientserver

Finger practice: simple HTTP client/server sending/receiving JSON. The server allows posting JSON and then responds with the same JSON parsed and then serialized. The client sends a total amount of requests in batches, checks if the JSON returned is the same as the one sent and collects errors for final report.

## Usage

Run the server with `go run ./cmd/server`, then run the client with `go run ./cmd/client`.
See `-help` for both commands for options.
