.PHONY: dev build codegen all clean
.DEFAULT: build

build:
	CGO_ENABLED=0 go build -o build/server server.go

dev:
	go run server.go

codegen:
	go run github.com/99designs/gqlgen generate

all: codegen build

clean:
	rm -r build
