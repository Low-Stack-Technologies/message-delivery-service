package delivery

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/Low-Stack-Technologies/message-delivery-service/internal/config"
)

func TestEmailProvider_Send_SubjectEncoding(t *testing.T) {
	// 1. Setup Mock SMTP Server
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer l.Close()

	port := l.Addr().(*net.TCPAddr).Port

	// Channel to receive the message content
	msgChan := make(chan string, 1)

	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		writer := bufio.NewWriter(conn)

		// Simple SMTP Handshake
		writer.WriteString("220 mock.smtp.server ESMTP\r\n")
		writer.Flush()

		// Read HELO/EHLO
		reader.ReadString('\n')
		writer.WriteString("250-Hello\r\n")
		writer.WriteString("250 AUTH PLAIN\r\n")
		writer.Flush()

		// Read AUTH
		reader.ReadString('\n')
		writer.WriteString("235 Authentication successful\r\n")
		writer.Flush()

		// Read MAIL FROM
		reader.ReadString('\n')
		writer.WriteString("250 OK\r\n")
		writer.Flush()

		// Read RCPT TO
		reader.ReadString('\n')
		writer.WriteString("250 OK\r\n")
		writer.Flush()

		// Read DATA
		reader.ReadString('\n')
		writer.WriteString("354 End data with <CR><LF>.<CR><LF>\r\n")
		writer.Flush()

		// Read Body
		var bodyBuilder strings.Builder
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			if line == ".\r\n" {
				break
			}
			bodyBuilder.WriteString(line)
		}
		msgChan <- bodyBuilder.String()

		writer.WriteString("250 OK\r\n")
		writer.Flush()

		// Read QUIT
		reader.ReadString('\n')
		writer.WriteString("221 Bye\r\n")
		writer.Flush()
	}()

	// 2. Setup Config
	cfg := &config.Config{
		EmailAccounts: []config.EmailAccountConfig{
			{
				Address: "test@example.com",
				SMTP: struct {
					Host     string `yaml:"host"`
					Port     int    `yaml:"port"`
					Username string `yaml:"username"`
					Password string `yaml:"password"`
				}{
					Host:     "127.0.0.1",
					Port:     port,
					Username: "user",
					Password: "password",
				},
			},
		},
	}

	// 3. Send Email
	provider := NewEmailProvider(cfg)
	subject := "Test ÅÄÖ Subject"
	err = provider.Send("test@example.com", []string{"recipient@example.com"}, subject, "Body content", false)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	// 4. Verify Content
	select {
	case msg := <-msgChan:
		// We expect the subject to be encoded
		if strings.Contains(msg, "Subject: Test ÅÄÖ Subject") {
			t.Logf("Received message with raw subject:\n%s", msg)
			t.Fatal("Subject was not encoded")
		}

		// Check if it looks like an encoded word
		// Q-encoding uses "=?utf-8?q?..."
		if !strings.Contains(msg, "Subject: =?utf-8?q?") {
			t.Logf("Received message:\n%s", msg)
			t.Fatal("Subject does not appear to be encoded correctly")
		}

	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for email content")
	}
}
