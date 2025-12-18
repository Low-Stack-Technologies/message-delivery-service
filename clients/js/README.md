# Message Delivery Service TypeScript Client

A TypeScript/JavaScript client library for the Message Delivery Service, featuring automatic Ed25519 request signing and type-safe API wrappers.

## Installation

```bash
# Using Bun
bun add @lowstacktechnologies/mds-client

# Using npm
npm install @lowstacktechnologies/mds-client

# Using pnpm
pnpm add @lowstacktechnologies/mds-client
```

## Usage

```typescript
import { MdsClient } from "@lowstacktechnologies/mds-client";

// Initialize the client
const client = new MdsClient(
  "http://localhost:3000",  // Server URL
  "your-client-id",         // Client ID
  "your-private-key-base64" // Ed25519 private key (Base64 or OpenSSH format)
);

// Send an email
const response = await client.sendEmail({
  from: { address: "sender@example.com", name: "Sender Name" },
  to: "recipient@example.com",
  subject: "Hello World",
  content: { body: "This is the email body", isHtml: false }
});

console.log(response.message); // "Email accepted for delivery"

// Send an SMS
const smsResponse = await client.sendSms({
  senderName: "MyApp",
  to: "+1234567890",
  content: { body: "Hello from MDS!" }
});
```

## Key Management

The service uses Ed25519 signatures for authentication. You need a key pair to sign requests.

The client supports multiple private key formats:
- **Base64**: Raw 64-byte Ed25519 private key or 32-byte seed
- **OpenSSH**: Keys generated with `ssh-keygen -t ed25519`
- **Hex**: 64 or 128 character hex strings

### 1. Generate an Ed25519 Key Pair
The easiest way is using `ssh-keygen`:
```bash
ssh-keygen -t ed25519 -f mds_key -N ""
```
This generates:
- `mds_key`: Your **Private Key** (OpenSSH format).
- `mds_key.pub`: Your **Public Key** (OpenSSH format).

### 2. Prepare the Keys for Configuration
- **Public Key**: You must Base64 encode the public key before adding it to the `services` section of the Message Delivery Service `config.yaml`:
  ```bash
  cat mds_key.pub | base64 -w 0
  ```
- **Private Key**: The client accepts the content of `mds_key` as a string. You can also Base64 encode it if you prefer to store it that way:
  ```bash
  cat mds_key | base64 -w 0
  ```

## CLI Test Utility

The package includes an interactive CLI for testing:

```bash
# Run from cloned repository
bun run cli

# Run directly with Bun (via npm)
bunx @lowstacktechnologies/mds-client

# Or if installed globally
mds-cli
```

The CLI will prompt you for server URL, credentials, and message details.

## API Reference

### `MdsClient`

#### Constructor
```typescript
new MdsClient(serverUrl: string, clientId: string, privateKey: string | Uint8Array)
```

#### Methods

- `sendEmail(request: EmailRequest): Promise<SuccessResponse>`
- `sendSms(request: SmsRequest): Promise<SmsSuccessResponse>`
- `health(): Promise<{ status: string; timestamp: string }>`

## Types

```typescript
interface EmailRequest {
  from: { address: string; name?: string };
  to: string | EmailContact | (string | EmailContact)[];
  subject: string;
  content?: { body: string; isHtml?: boolean } | { template: { name: string; data: Record<string, unknown> } };
}

interface SmsRequest {
  senderName: string;
  to: string | { phone: string; country: string } | (string | { phone: string; country: string })[];
  content?: { body: string } | { template: { name: string; data: Record<string, unknown> } };
}
```
