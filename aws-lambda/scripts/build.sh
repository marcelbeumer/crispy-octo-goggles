#!/bin/sh
GOOS=linux GOARCH=amd64 go build -o dist/main main.go
