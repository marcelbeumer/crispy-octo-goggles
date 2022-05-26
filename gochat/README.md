# GoChat

Basic chat room application using WebSockets, gRPC and a terminal UI. Requires Go 1.18+.

<img src="./assets/screenshot.jpg" width="800px" />

## Installation

### Go install

```
go install -a github.com/marcelbeumer/crispy-octo-goggles/gochat@latest
```

### Docker

```
docker build -t gochat .
docker run -it --rm --network host gochat
```

## Run from source

```
go run .
```

## Usage

To run a server:

```
gochat server -V # websockets
gochat server -V --grpc # gRPC
```

To connect a client:

```
gochat client -u Mario # websockets
gochat client -u Mario --grpc # gRPC
```

For more options and details see:

```
gochat --help
```

## TODO

- Try to use [copygen](https://github.com/switchupcb/copygen) for mapping gRPC structs.
- Write more tests.
