#!/bin/bash
# Generate Go code from OpenAPI specification using oapi-codegen

# Ensure the output directory exists
mkdir -p service/pkg/api

# Run oapi-codegen
# -package api: Set the Go package name
# -generate types,chi-server: Generate model types and chi router boilerplate
# -o service/pkg/api/api.gen.go: Output file path
oapi-codegen -package api -generate types,chi-server -o service/pkg/api/api.gen.go openapi.yaml

echo "Successfully generated types in service/pkg/api/api.gen.go"

# Generate Go Client
mkdir -p clients/go/api
oapi-codegen -package api -generate types,client -o clients/go/api/client.gen.go openapi.yaml
echo "Successfully generated Go client in clients/go/api/client.gen.go"

# Generate TS Client Types
cd clients/js
bunx openapi-typescript ../../openapi.yaml -o src/schema.ts
echo "Successfully generated TS types in clients/js/src/schema.ts"
cd ../..
