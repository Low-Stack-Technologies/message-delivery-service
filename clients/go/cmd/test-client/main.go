package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/esaiaswestberg/message-delivery-service/clients/go"
	"github.com/esaiaswestberg/message-delivery-service/clients/go/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("--- Message Delivery Service Test Utility ---")

	serverURL := prompt(reader, "Server URL (e.g., http://localhost:3000)")
	clientID := prompt(reader, "Client ID")
	privKey := prompt(reader, "Private Key (Base64)")

	client, err := mds.NewClient(serverURL, clientID, privKey)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	fmt.Println("\nChoose message type:")
	fmt.Println("1. Email")
	fmt.Println("2. SMS")
	choice := prompt(reader, "Selection (1 or 2)")

	ctx := context.Background()

	switch choice {
	case "1":
		from := prompt(reader, "Sender Email")
		to := prompt(reader, "Recipient Email")
		subject := prompt(reader, "Subject")
		body := prompt(reader, "Body")

		req := api.EmailRequest{
			From:    api.EmailContact{Address: openapi_types.Email(from)},
			Subject: subject,
		}
		req.To.FromEmailRequestTo0(openapi_types.Email(to))
		
		content := api.EmailRequest_Content{}
		content.FromEmailRequestContent0(api.EmailRequestContent0{
			Body: body,
		})
		req.Content = &content

		fmt.Println("\nSending Email...")
		resp, err := client.SendEmail(ctx, req)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Printf("Success! Message: %s\n", resp.Message)

	case "2":
		from := prompt(reader, "Sender Name (3-11 chars)")
		to := prompt(reader, "Recipient Phone (E.164)")
		body := prompt(reader, "Message Body")

		req := api.SmsRequest{
			SenderName: from,
		}
		
		recipient := api.SmsRecipient{}
		recipient.FromSmsRecipient0(to)
		req.To.FromSmsRecipient(recipient)

		content := api.SmsRequest_Content{}
		content.FromSmsRequestContent0(api.SmsRequestContent0{
			Body: body,
		})
		req.Content = &content

		fmt.Println("\nSending SMS...")
		resp, err := client.SendSms(ctx, req)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Printf("Success! Message: %s\n", resp.Message)

	default:
		fmt.Println("Invalid choice.")
	}
}

func prompt(reader *bufio.Reader, label string) string {
	fmt.Printf("%s: ", label)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
