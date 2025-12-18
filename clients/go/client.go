package mds

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Low-Stack-Technologies/message-delivery-service/clients/go/api"
	"golang.org/x/crypto/ssh"
)

// Client is a high-level wrapper around the Message Delivery Service API.
type Client struct {
	apiClient *api.ClientWithResponses
	clientID  string
	privKey   ed25519.PrivateKey
}

// NewClient creates a new Message Delivery Service client.
// privKeyB64 can be:
// 1. A raw 64-byte Ed25519 private key (Base64)
// 2. A 32-byte Ed25519 seed (Base64)
// 3. A PEM-encoded PKCS8 private key (either raw PEM string or Base64 of it)
// 4. An OpenSSH private key (either raw PEM string or Base64 of it)
func NewClient(serverURL, clientID, privKeyStr string) (*Client, error) {
	var privKeyBytes []byte
	var err error

	// Try decoding as raw Base64 first
	privKeyBytes, err = base64.StdEncoding.DecodeString(privKeyStr)
	if err != nil {
		// If not valid base64, it might be a raw PEM string
		privKeyBytes = []byte(privKeyStr)
	}

	var privKey ed25519.PrivateKey

	// Handle PEM/OpenSSH encoding
	if bytes.Contains(privKeyBytes, []byte("BEGIN")) {
		// Try OpenSSH format first
		key, err := ssh.ParseRawPrivateKey(privKeyBytes)
		if err == nil {
			if pk, ok := key.(*ed25519.PrivateKey); ok {
				privKey = *pk
			} else if pk, ok := key.(ed25519.PrivateKey); ok {
				privKey = pk
			} else {
				return nil, fmt.Errorf("key is not an Ed25519 private key (got %T)", key)
			}
		} else {
			// Try PKCS8
			block, _ := pem.Decode(privKeyBytes)
			if block != nil {
				key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
				if err != nil {
					return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
				}
				var ok bool
				privKey, ok = key.(ed25519.PrivateKey)
				if !ok {
					return nil, fmt.Errorf("key is not an Ed25519 private key (PKCS8)")
				}
			} else {
				return nil, fmt.Errorf("failed to parse PEM: %w", err)
			}
		}
	} else {
		// Handle raw bytes
		switch len(privKeyBytes) {
		case ed25519.PrivateKeySize: // 64 bytes
			privKey = ed25519.PrivateKey(privKeyBytes)
		case ed25519.SeedSize: // 32 bytes
			privKey = ed25519.NewKeyFromSeed(privKeyBytes)
		default:
			return nil, fmt.Errorf("invalid private key size: expected 32 (seed), 64 (private), or PEM block, got %d bytes", len(privKeyBytes))
		}
	}

	c := &Client{
		clientID: clientID,
		privKey:  privKey,
	}

	apiClient, err := api.NewClientWithResponses(serverURL, api.WithRequestEditorFn(c.signerInterceptor))
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	c.apiClient = apiClient
	return c, nil
}

// signerInterceptor is an oapi-codegen RequestEditorFn that automatically signs requests.
func (c *Client) signerInterceptor(ctx context.Context, req *http.Request) error {
	timestamp := time.Now().Format(time.RFC3339)

	// 1. Read Body
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("failed to read request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// 2. Construct Canonical Request
	bodyHash := sha256.Sum256(bodyBytes)
	bodyHashHex := hex.EncodeToString(bodyHash[:])

	// Canonical = Method + "\n" + Path + "\n" + X-Timestamp + "\n" + SHA256(Body)
	canonical := req.Method + "\n" + req.URL.Path + "\n" + timestamp + "\n" + bodyHashHex

	// 3. Sign
	signature := ed25519.Sign(c.privKey, []byte(canonical))
	signatureB64 := base64.StdEncoding.EncodeToString(signature)

	// 4. Set Headers
	req.Header.Set("X-Client-Id", c.clientID)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("Authorization", "Signature "+signatureB64)

	return nil
}

// SendEmail sends an email request to the service.
func (c *Client) SendEmail(ctx context.Context, emailReq api.EmailRequest) (*api.SuccessResponse, error) {
	resp, err := c.apiClient.PostV3EmailWithResponse(ctx, &api.PostV3EmailParams{
		XClientId:  c.clientID,
		XTimestamp: time.Now(),
	}, emailReq)
	if err != nil {
		return nil, err
	}

	if resp.JSON202 != nil {
		return resp.JSON202, nil
	}

	// Fallback for 202 Accepted if JSON202 is nil (e.g. Content-Type mismatch)
	if resp.StatusCode() == http.StatusAccepted {
		var successResp api.SuccessResponse
		if err := json.Unmarshal(resp.Body, &successResp); err == nil {
			return &successResp, nil
		}
		// If unmarshal fails but it's 202, still consider it success
		return &api.SuccessResponse{Success: true, Message: "Accepted (raw)"}, nil
	}

	// Try parsing as ErrorResponse
	var errResp api.ErrorResponse
	if err := json.Unmarshal(resp.Body, &errResp); err == nil && !errResp.Success && errResp.Error.Code != "" {
		return nil, fmt.Errorf("API error (%s): %s", errResp.Error.Code, errResp.Error.Message)
	}

	return nil, fmt.Errorf("API error: %s", resp.Status())
}

// SendSms sends an SMS request to the service.
func (c *Client) SendSms(ctx context.Context, smsReq api.SmsRequest) (*api.SmsSuccessResponse, error) {
	resp, err := c.apiClient.PostV3SmsWithResponse(ctx, &api.PostV3SmsParams{
		XClientId:  c.clientID,
		XTimestamp: time.Now(),
	}, smsReq)
	if err != nil {
		return nil, err
	}

	if resp.JSON202 != nil {
		return resp.JSON202, nil
	}

	// Fallback for 202 Accepted if JSON202 is nil (e.g. Content-Type mismatch)
	if resp.StatusCode() == http.StatusAccepted {
		var successResp api.SmsSuccessResponse
		if err := json.Unmarshal(resp.Body, &successResp); err == nil {
			return &successResp, nil
		}
		// If unmarshal fails but it's 202, still consider it success
		return &api.SmsSuccessResponse{Success: true, Message: "Accepted (raw)"}, nil
	}

	// Try parsing as ErrorResponse
	var errResp api.ErrorResponse
	if err := json.Unmarshal(resp.Body, &errResp); err == nil && !errResp.Success && errResp.Error.Code != "" {
		return nil, fmt.Errorf("API error (%s): %s", errResp.Error.Code, errResp.Error.Message)
	}

	return nil, fmt.Errorf("API error: %s", resp.Status())
}
