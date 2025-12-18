# Message Delivery Service

A high-performance, secure, and extensible service for delivering transactional emails and SMS messages. Built for stability and ease of integration.

## Features

- **Multi-Provider Support**: Pluggable backends for email (SMTP) and SMS (46elks).
- **Request Signing**: Ed25519 asymmetric signatures for top-tier security.
- **Hot Reload**: Live configuration updates without service downtime.
- **Type-Safe API**: Fully documented via OpenAPI 3.0.
- **Resilient Delivery**: Robust handling of SMTP implicit SSL/TLS and batch operations.

## Project Structure

- **[`/service`](./service)**: The core Go-based delivery service. Handles routing, authentication, and provider orchestration.
- **[`/clients/go`](./clients/go)**: Official Go client library with automatic request signing.
- **[`/openapi.yaml`](./openapi.yaml)**: The source-of-truth API specification.
- **[`/generate.sh`](./generate.sh)**: Utility script to synchronize code generators with the OpenAPI spec.

## Getting Started

1. **Deploy the Service**: Check the [Service README](./service/README.md) for Docker Compose setup.
2. **Configure Backends**: Add your SMTP and SMS provider credentials to `config.yaml`.
3. **Integrate your App**: Use the [Go Client](./clients/go/README.md) for seamless integration.

## Development

To regenerate API types after modifying `openapi.yaml`:
```bash
./generate.sh
```
