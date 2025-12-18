package delivery

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Low-Stack-Technologies/message-delivery-service/internal/config"
)

type SmsProvider struct {
	config config.FortySixElksConfig
}

func NewSmsProvider(cfg *config.Config) *SmsProvider {
	return &SmsProvider{config: cfg.Sms.FortySixElks}
}

func (p *SmsProvider) Send(from string, to []string, body string) error {
	// 46elks implementation
	apiURL := "https://api.46elks.com/a1/sms"

	config.DebugLog("[DEBUG] SMS Delivery - Sending to %d recipients via 46elks", len(to))
	for _, recipient := range to {
		config.DebugLog("[DEBUG] SMS Delivery - Recipient: %s", recipient)
		data := url.Values{}
		data.Set("from", from)
		data.Set("to", recipient)
		data.Set("message", body)

		req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
		if err != nil {
			return err
		}

		req.SetBasicAuth(p.config.Username, p.config.Password)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			config.DebugLog("[DEBUG] SMS Delivery Failed - 46elks error: %s", resp.Status)
			return fmt.Errorf("46elks API error: %s", resp.Status)
		}
		config.DebugLog("[DEBUG] SMS Delivery Success - Sent to %s", recipient)
	}

	return nil
}
