#!/bin/sh
set -e
export APOLLO_ELV2_LICENSE=accept
./scripts/compose-supergraph.sh
./bin/router --supergraph ./generated/supergraph.graphql -c ./conf/router.prod.yaml
