#!/bin/sh
set -e
./scripts/build.sh
./scripts/zip.sh
./scripts/upload.sh
