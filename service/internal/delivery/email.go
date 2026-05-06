package delivery

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net/smtp"

	"github.com/Low-Stack-Technologies/message-delivery-service/internal/config"
)

type EmailProvider struct {
	cfg *config.Config
}

func NewEmailProvider(cfg *config.Config) *EmailProvider {
	return &EmailProvider{cfg: cfg}
}

func (p *EmailProvider) Send(from string, to []string, subject string, body string, isHTML bool) error {
	cfg := config.Get()
	if cfg == nil {
		cfg = p.cfg
	}
	if cfg == nil {
		return fmt.Errorf("configuration snapshot is not available")
	}
	var acc *config.EmailAccountConfig
	for i := range cfg.EmailAccounts {
		if cfg.EmailAccounts[i].Address == from {
			acc = &cfg.EmailAccounts[i]
			break
		}
	}
	if acc == nil {
		for i := range cfg.EmailAccounts {
			if cfg.EmailAccounts[i].IsDefault {
				acc = &cfg.EmailAccounts[i]
				break
			}
		}
	}

	if acc == nil {
		config.DebugLog("[DEBUG] Email Delivery Failed - No account for: %s", from)
		return fmt.Errorf("no SMTP account configured for sender: %s", from)
	}

	if from == "" {
		from = acc.Address
	}

	config.DebugLog("[DEBUG] Email Delivery - Using SMTP account: %s (%s:%d)", acc.Address, acc.SMTP.Host, acc.SMTP.Port)

	contentType := "text/plain"
	if isHTML {
		contentType = "text/html"
	}
	msg := buildEmailMessage(from, to, subject, body, contentType)

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

		if err = client.Mail(from); err != nil {
			return err
		}
		for _, addr := range to {
			if err = client.Rcpt(addr); err != nil {
				return err
			}
		}

		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(msg))
		if err != nil {
			return err
		}
		err = w.Close()
		if err != nil {
			return err
		}

		return client.Quit()
	}

	if err := smtp.SendMail(addr, auth, from, to, []byte(msg)); err != nil {
		config.DebugLog("[DEBUG] Email Delivery Failed - SMTP Error: %v", err)
		return err
	}
	config.DebugLog("[DEBUG] Email Delivery Success - Sent to %v", to)
	return nil
}

func buildEmailMessage(from string, to []string, subject string, body string, contentType string) string {
	encodedSubject := mime.QEncoding.Encode("utf-8", subject)
	return fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: %s; charset=\"UTF-8\"\r\n"+
		"\r\n"+
		"%s", from, to[0], encodedSubject, contentType, body)
}
