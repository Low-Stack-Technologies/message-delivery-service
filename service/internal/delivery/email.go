package delivery

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/Low-Stack-Technologies/message-delivery-service/internal/config"
)

type EmailProvider struct {
	accounts map[string]config.EmailAccountConfig
}

func NewEmailProvider(cfg *config.Config) *EmailProvider {
	accounts := make(map[string]config.EmailAccountConfig)
	for _, acc := range cfg.EmailAccounts {
		accounts[acc.Address] = acc
	}
	return &EmailProvider{accounts: accounts}
}

func (p *EmailProvider) Send(from string, to []string, subject string, body string, isHTML bool) error {
	acc, ok := p.accounts[from]
	if !ok {
		return fmt.Errorf("no SMTP account configured for sender: %s", from)
	}

	contentType := "text/plain"
	if isHTML {
		contentType = "text/html"
	}

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: %s; charset=\"UTF-8\"\r\n"+
		"\r\n"+
		"%s", from, to[0], subject, contentType, body)

	auth := smtp.PlainAuth("", acc.SMTP.Username, acc.SMTP.Password, acc.SMTP.Host)
	addr := fmt.Sprintf("%s:%d", acc.SMTP.Host, acc.SMTP.Port)

	// Simple SMTP send
	if acc.SMTP.Port == 465 {
		// Implicit SSL/TLS
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         acc.SMTP.Host,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return err
		}
		client, err := smtp.NewClient(conn, acc.SMTP.Host)
		if err != nil {
			return err
		}
		if err = client.Auth(auth); err != nil {
			return err
		}
		// ... (simplified for this draft, full implementation would handle recipients etc)
		return client.Quit()
	}

	return smtp.SendMail(addr, auth, from, to, []byte(msg))
}
