#!/bin/sh
curl -sSL https://rover.apollo.dev/nix/latest | sh
mkdir -p bin
cd bin
curl -sSL https://router.apollo.dev/download/nix/latest | sh
