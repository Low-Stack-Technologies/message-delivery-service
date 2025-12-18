import * as ed from "@noble/ed25519";
import { createHash } from "crypto";
import type {
  EmailRequest,
  SmsRequest,
  SuccessResponse,
  SmsSuccessResponse,
  ErrorResponse,
} from "./types.js";

// Configure noble/ed25519 to use native crypto SHA512
ed.etc.sha512Sync = (...messages: Uint8Array[]) => {
  const hash = createHash("sha512");
  for (const msg of messages) {
    hash.update(msg);
  }
  return new Uint8Array(hash.digest());
};

export * from "./types.js";

/**
 * MdsClient is a high-level wrapper around the Message Delivery Service API.
 */
export class MdsClient {
  private serverUrl: string;
  private clientId: string;
  private privateKey: Uint8Array;

  /**
   * Creates a new Message Delivery Service client.
   * @param serverUrl - The base URL of the MDS server (e.g., "http://localhost:3000")
   * @param clientId - Your registered client ID
   * @param privateKey - Ed25519 private key as Base64 string, hex string, or Uint8Array (64 bytes or 32-byte seed)
   */
  constructor(serverUrl: string, clientId: string, privateKey: string | Uint8Array) {
    this.serverUrl = serverUrl.replace(/\/$/, ""); // Remove trailing slash
    this.clientId = clientId;
    this.privateKey = this.parsePrivateKey(privateKey);
  }

  private parsePrivateKey(key: string | Uint8Array): Uint8Array {
    if (key instanceof Uint8Array) {
      return this.normalizeKeyBytes(key);
    }

    // Try Base64 decode first
    try {
      const decoded = Buffer.from(key, "base64");
      
      // Check if the decoded content is actually a PEM file (OpenSSH)
      const decodedStr = decoded.toString();
      if (decodedStr.includes("BEGIN OPENSSH PRIVATE KEY")) {
        return this.parseOpenSSHKey(decodedStr);
      }

      if (decoded.length === 32 || decoded.length === 64) {
        return this.normalizeKeyBytes(decoded);
      }
    } catch {
      // Not Base64 or invalid, continue
    }

    // Try OpenSSH format
    if (key.includes("BEGIN OPENSSH PRIVATE KEY")) {
      return this.parseOpenSSHKey(key);
    }

    // Try hex decode
    if (/^[0-9a-fA-F]+$/.test(key) && (key.length === 64 || key.length === 128)) {
      const decoded = Buffer.from(key, "hex");
      return this.normalizeKeyBytes(decoded);
    }

    throw new Error(
      "Invalid private key format. Expected Base64, hex, OpenSSH, or raw bytes (32 or 64 bytes)"
    );
  }

  private normalizeKeyBytes(bytes: Uint8Array | Buffer): Uint8Array {
    const arr = new Uint8Array(bytes);
    if (arr.length === 64) {
      // Full private key (seed + public key), extract seed
      return arr.slice(0, 32);
    }
    if (arr.length === 32) {
      return arr;
    }
    throw new Error(`Invalid key size: expected 32 or 64 bytes, got ${arr.length}`);
  }

  private parseOpenSSHKey(pemContent: string): Uint8Array {
    // OpenSSH format: extract the base64 body between BEGIN and END markers
    const lines = pemContent.split("\n");
    const bodyLines = lines.filter(
      (line) => !line.startsWith("-----") && line.trim().length > 0
    );
    const body = Buffer.from(bodyLines.join(""), "base64");

    // OpenSSH Ed25519 keys have a specific structure
    // The private key seed is located after the public key section
    // For unencrypted keys, we can find the 64-byte private key at a known offset
    // The format is: "openssh-key-v1\0" + padding + ... + privkey (64 bytes)
    
    // Simple approach: scan for the 64-byte private key section
    // In unencrypted Ed25519 keys, the private key appears after the public key
    const magicPrefix = Buffer.from("openssh-key-v1\0");
    if (!body.subarray(0, magicPrefix.length).equals(magicPrefix)) {
      throw new Error("Invalid OpenSSH key format");
    }

    // The 64-byte private key (seed + pubkey) is near the end
    // We need to find it by parsing the structure
    // For simplicity, we'll look for a 64-byte sequence followed by a 32-byte pubkey
    // that matches the pubkey in the private section
    
    // Quick approach: the last 32 bytes before the comment are the public key part
    // The 32 bytes before that are the seed
    // This is a simplification that works for unencrypted keys
    
    // Find the ed25519 key data by searching for the key type string
    const keyTypeStr = "ssh-ed25519";
    let idx = body.indexOf(keyTypeStr, magicPrefix.length);
    if (idx === -1) {
      throw new Error("Could not find Ed25519 key type in OpenSSH key");
    }

    // Skip to after the public key section to find the private key
    // The private section has: 2x uint32 check bytes, key type, pubkey, privkey (64 bytes), comment
    idx = body.indexOf(keyTypeStr, idx + keyTypeStr.length);
    if (idx === -1) {
      throw new Error("Could not find private key section in OpenSSH key");
    }

    // Move past key type length (4 bytes) + key type + pubkey length (4 bytes) + pubkey (32 bytes)
    idx += keyTypeStr.length;
    idx += 4 + 32; // pubkey length + pubkey

    // Now we should be at privkey length (4 bytes) followed by 64-byte privkey
    const privkeyLen = body.readUInt32BE(idx);
    idx += 4;

    if (privkeyLen !== 64) {
      throw new Error(`Unexpected private key length: ${privkeyLen}`);
    }

    const fullPrivKey = body.subarray(idx, idx + 64);
    return new Uint8Array(fullPrivKey.subarray(0, 32)); // Return just the seed
  }

  private async signRequest(method: string, path: string, body: string): Promise<{
    timestamp: string;
    signature: string;
  }> {
    const timestamp = new Date().toISOString();

    // SHA256 hash of body
    const encoder = new TextEncoder();
    const bodyBytes = encoder.encode(body);
    const hashBuffer = await crypto.subtle.digest("SHA-256", bodyBytes);
    const bodyHash = Buffer.from(hashBuffer).toString("hex");

    // Canonical request: Method + "\n" + Path + "\n" + Timestamp + "\n" + SHA256(Body)
    const canonical = `${method}\n${path}\n${timestamp}\n${bodyHash}`;
    const canonicalBytes = encoder.encode(canonical);

    // Sign with Ed25519
    const signature = ed.sign(canonicalBytes, this.privateKey);
    const signatureB64 = Buffer.from(signature).toString("base64");

    return { timestamp, signature: signatureB64 };
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown
  ): Promise<T> {
    const bodyStr = body ? JSON.stringify(body) : "";
    const { timestamp, signature } = await this.signRequest(method, path, bodyStr);

    const response = await fetch(`${this.serverUrl}${path}`, {
      method,
      headers: {
        "Content-Type": "application/json",
        "X-Client-Id": this.clientId,
        "X-Timestamp": timestamp,
        Authorization: `Signature ${signature}`,
      },
      body: bodyStr || undefined,
    });

    const responseBody = await response.text();
    let parsed: unknown;
    try {
      parsed = JSON.parse(responseBody);
    } catch {
      if (response.status === 202) {
        return { success: true, message: "Accepted" } as T;
      }
      throw new Error(`API error: ${response.status} ${response.statusText}`);
    }

    if (response.status === 202 || response.status === 200) {
      return parsed as T;
    }

    const errorResp = parsed as ErrorResponse;
    if (errorResp?.error?.code) {
      throw new Error(`API error (${errorResp.error.code}): ${errorResp.error.message}`);
    }

    throw new Error(`API error: ${response.status} ${response.statusText}`);
  }

  /**
   * Sends an email through the Message Delivery Service.
   */
  async sendEmail(request: EmailRequest): Promise<SuccessResponse> {
    return this.request<SuccessResponse>("POST", "/v3/email", request);
  }

  /**
   * Sends an SMS through the Message Delivery Service.
   */
  async sendSms(request: SmsRequest): Promise<SmsSuccessResponse> {
    return this.request<SmsSuccessResponse>("POST", "/v3/sms", request);
  }

  /**
   * Checks the health of the Message Delivery Service.
   */
  async health(): Promise<{ status: string; timestamp: string }> {
    const response = await fetch(`${this.serverUrl}/health`);
    if (!response.ok) {
      throw new Error(`Health check failed: ${response.status} ${response.statusText}`);
    }
    const text = await response.text();
    try {
      return JSON.parse(text);
    } catch {
      throw new Error(`Failed to parse JSON response: "${text.substring(0, 100)}..."`);
    }
  }
}
