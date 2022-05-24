# GoChat

Basic chat room application on websockets or gRPC.

## Installation

Install the `gochat` binary with:

```
go install github.com/marcelbeumer/crispy-octo-goggles/gochat@latest
```

## Running from source

Run directly from source with

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

- Write more tests.
- Production setup: Dockerfile, (client) install/usage instructions.
