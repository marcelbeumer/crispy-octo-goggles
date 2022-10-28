#!/bin/bash
set -e
source scripts/dotenv.sh
.env export
./scripts/compose-supergraph.sh
./bin/router --supergraph ./generated/supergraph.graphql -c ./conf/router.dev.yaml
