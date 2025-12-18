# Message Delivery Service Go Client

A robust Go client library for interacting with the Message Delivery Service, featuring automatic Ed25519 request signing and type-safe API wrappers.

## Table of Contents
- [Installation](#installation)
- [Usage](#usage)
- [Key Management](#key-management)
- [Interactive Test Client](#interactive-test-client)

## Installation

```bash
go get github.com/esaiaswestberg/message-delivery-service/clients/go
```

## Usage

```go
package main

import (
	"context"
	"log"

	mds "github.com/esaiaswestberg/message-delivery-service/clients/go"
	"github.com/esaiaswestberg/message-delivery-service/clients/go/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func main() {
	// 1. Initialize Client
	// Supports raw Base64, 32-byte seed, PKCS8 PEM, or OpenSSH formats.
	client, err := mds.NewClient("http://localhost:3000", "your-client-id", "your-private-key-string")
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// 2. Send Email
	req := api.EmailRequest{
		From:    api.EmailContact{Address: openapi_types.Email("sender@example.com")},
		Subject: "Hello World",
	}
	// ... (configure recipients and content)

	resp, err := client.SendEmail(context.Background(), req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Printf("Success! Message: %s", resp.Message)
}
```

## Key Management

The service uses Ed25519 signatures for authentication. You need a key pair to sign requests.

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

## Interactive Test Client

We provide a CLI utility to test your connectivity and credentials.

### 1. Generate API Types
If you are developing inside the repository, ensure types are generated first:
```bash
# From the project root
./generate.sh
```

### 2. Run the Test Client
```bash
cd clients/go
go run cmd/test-client/main.go
```
The utility will prompt you for your Server URL, Client ID, and Private Key, and then allow you to send a test Email or SMS.
