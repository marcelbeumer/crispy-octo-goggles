#!/bin/sh
curl -sSL https://raw.githubusercontent.com/apollographql/rover/v0.9.1/installers/binstall/scripts/nix/install.sh | sh
mkdir -p bin
cd bin
curl -sSL https://raw.githubusercontent.com/apollographql/router/v1.2.1/scripts/install.sh | sh
