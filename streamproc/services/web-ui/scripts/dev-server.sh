#!/bin/sh
cd ./dev-server
go run . -s ../public "${@}"
