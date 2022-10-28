# services/router

## Local development

- Run subgraph services in the background.
- Copy `.env.example` to `.env` (modify if needed).
- Run `./scripts/install.sh` to install Apollo Router and Rover.
- Run `./scripts/run-dev.sh` to compose the supergraph and run the router.

## Running docker / production

- Build docker image using `Dockerfile`.
- Run image with env vars also found in `.env.example`:
  - `COMMERCE_URL=<url>`
  - `CONTENT_URL=<url>`
