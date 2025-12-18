# Message Delivery Service

This is a service to easily provide email and SMS sending capabilities to other services through a simple REST API. It is built using Node.js and Express along with Zod for validation.

# Docker Compose

The easiest way to run this service is through Docker Compose. The following is an example `docker-compose.yml` file:

```yaml
services:
  message-delivery-service:
    image: dotzerotechnologies/message-delivery-service:latest
    ports:
      - 3000:3000
    environment:
      - DATA_PATH=/data
      - CONFIGURATION_FILE_PATH=/data/config.json
    volumes:
      - ./data:/data
    restart: unless-stopped
```

# Documentation

The API documentation is available in the [OpenAPI specification](openapi.yaml).
