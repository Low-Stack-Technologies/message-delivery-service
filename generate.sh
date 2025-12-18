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
