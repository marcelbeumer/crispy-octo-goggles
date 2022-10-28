#!/bin/bash
cat <<EOF | ~/.rover/bin/rover supergraph compose --config - > generated/supergraph.graphql
federation_version: 2
subgraphs:
  content:
    routing_url: $CONTENT_URL
    schema:
      subgraph_url: $CONTENT_URL
  commerce:
    routing_url: $COMMERCE_URL
    schema:
      subgraph_url: $COMMERCE_URL
EOF
