# Snippetbox

Example app from the book [Let's Go](https://lets-go.alexedwards.net).

- Run mysql with `docker compose up -d`.
- Setup up database with `./scripts/setup_db.sh`.
- Generate TLS cert for `localhost` with `./scripts/generate_cert.sh localhost`.
- Start the server with `go run ./cmd/web`.
- Open `https://localhost:4000` (and accept self signed certificate)

## Ideas / TODO

- Implement PRG pattern for invalid forms (session pop string of state struct JSON for example)
