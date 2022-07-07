#!/bin/bash
set -e
HOST=$1

mkdir -p tls
cd tls
go run /usr/local/go/src/crypto/tls/generate_cert.go --rsa-bits=2048 --host=$HOST
