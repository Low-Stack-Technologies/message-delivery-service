# Message Delivery Service

The core backend service for orchestrating email and SMS delivery. It handles request authentication, provider selection, and delivery tracking.

## Deployment with Docker Compose

The service is available as a Docker image on `ghcr.io`.

### 1. Create a `docker-compose.yaml`
```yaml
services:
  mds:
    image: ghcr.io/low-stack-technologies/message-delivery-service:latest
    ports:
      - "3000:3000"
    volumes:
      - ./config.yaml:/app/config.yaml
    restart: unless-stopped
```

### 2. Initialize Configuration
Create a `config.yaml` file in the same directory. The service will generate a default one if it's missing, but you should configure your providers and authorized services.

## Configuration Guide

The `config.yaml` file is the central source of truth and supports **hot-reloading**.

### 1. General Settings
```yaml
debug: false # Set to true to enable verbose tracing in server logs
```

### 2. Authorized Services (Signature Auth)
Every client using the API must be registered here with their Ed25519 public key.
```yaml
services:
  - id: "my-app"
    name: "My Application"
    public_key: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA..." # OpenSSH or PKCS8 Base64
```

### 2. Email Configuration (SMTP)
You can add multiple SMTP accounts. The service selects the account based on the `from` address in the API request.
```yaml
email_accounts:
  - address: "support@example.com"
    smtp:
      host: "smtp.example.com"
      port: 465 # Supports 465 (Implicit SSL/TLS) and 587 (STARTTLS)
      username: "user@example.com"
      password: "your-password"
```

### 3. SMS Configuration (46elks)
```yaml
sms:
  46elks:
    username: "api_user_id"
    password: "api_password"
```

## Monitoring

- **Health Check**: `GET /health` (Public) - Returns 200 OK if the service is running.
- **Logs**: The service logs all authentication attempts and delivery statuses with `[DEBUG]` prefixes for easy troubleshooting.
