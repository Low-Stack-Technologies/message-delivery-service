# Message Delivery Service

The core backend service for orchestrating email and SMS delivery. It handles request authentication, provider selection, and delivery tracking.

## Deployment

The service persists all runtime configuration in SQLite.

### Runtime settings

- `MDS_DB_PATH` sets the SQLite file path. If omitted, the service uses `message-delivery-service.db` in the working directory.
- `--reset-db` removes the database file, recreates the schema, and exits.
- `--seed-db` removes the database file, recreates the schema, seeds the demo dataset, and exits.

### Default admin user

On first start the service creates a default admin user and logs the username, password, and TOTP provisioning details to stdout. That is the bootstrap credential for the admin web UI.

### Data model

The database stores:

- signed service identities
- SMTP accounts
- 46elks credentials
- outbound message history
- admin users and sessions
- activity log entries

### Docker Compose

If you run the service in Docker, mount the SQLite file or a data directory:

```yaml
services:
  mds:
    image: ghcr.io/low-stack-technologies/message-delivery-service:latest
    ports:
      - "3000:3000"
    environment:
      - MDS_DB_PATH=/data/message-delivery-service.db
    volumes:
      - ./data:/data
    restart: unless-stopped
```

## Monitoring

- **Health Check**: `GET /health` (Public) - Returns 200 OK if the service is running.
- **Logs**: The service logs default admin bootstrap credentials and delivery/authentication activity with `[DEBUG]` prefixes for troubleshooting.
