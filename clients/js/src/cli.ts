#!/usr/bin/env node
import { MdsClient } from "./index.js";
import * as readline from "readline";

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
});

function prompt(question: string): Promise<string> {
  return new Promise((resolve) => {
    rl.question(question, (answer) => {
      resolve(answer.trim());
    });
  });
}

async function main() {
  console.log("\n--- Message Delivery Service Test Utility (TypeScript) ---\n");

  const serverUrl = await prompt("Server URL (e.g., http://localhost:3000): ");
  const clientId = await prompt("Client ID: ");
  const privateKey = await prompt("Private Key (Base64): ");

  let client: MdsClient;
  try {
    client = new MdsClient(serverUrl, clientId, privateKey);
    console.log("\n✓ Client initialized successfully\n");
  } catch (error) {
    console.error("✗ Failed to initialize client:", (error as Error).message);
    rl.close();
    process.exit(1);
  }

  // Check health
  try {
    const health = await client.health();
    console.log(`✓ Server health: ${health.status}\n`);
  } catch (error) {
    console.error("✗ Health check failed:", (error as Error).message);
  }

  console.log("Choose message type:");
  console.log("1. Email");
  console.log("2. SMS");
  const choice = await prompt("Selection (1 or 2): ");

  if (choice === "1") {
    const senderEmail = await prompt("Sender Email: ");
    const recipientEmail = await prompt("Recipient Email: ");
    const subject = await prompt("Subject: ");
    const body = await prompt("Body: ");

    console.log("\nSending Email...");
    try {
      const response = await client.sendEmail({
        from: { address: senderEmail },
        to: recipientEmail,
        subject,
        content: { body, isHtml: false },
      });
      console.log("✓ Success!", response.message);
    } catch (error) {
      console.error("✗ Error:", (error as Error).message);
    }
  } else if (choice === "2") {
    const senderName = await prompt("Sender Name: ");
    const recipientPhone = await prompt("Recipient Phone (E.164): ");
    const body = await prompt("Body: ");

    console.log("\nSending SMS...");
    try {
      const response = await client.sendSms({
        senderName,
        to: recipientPhone,
        content: { body },
      });
      console.log("✓ Success!", response.message);
    } catch (error) {
      console.error("✗ Error:", (error as Error).message);
    }
  } else {
    console.log("Invalid selection");
  }

  rl.close();
}

main().catch((error) => {
  console.error("Fatal error:", error);
  rl.close();
  process.exit(1);
});
