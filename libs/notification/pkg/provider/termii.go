package provider

import (
	"os"

	ntxhttp "github.com/thescaffold/gox-packages/libs/core/http"
)

type TermiiProvider struct {
	client *ntxhttp.Client `inject:""`
}

func (p *TermiiProvider) Email(_ EmailMessage) bool { return false }

func (p *TermiiProvider) SMS(msg SMSMessage) bool {
	baseurl := os.Getenv("TERMII_BASE_URL")
	apiKey := os.Getenv("TERMII_API_KEY")
	senderID := os.Getenv("TERMII_SENDER_ID")

	from := senderID
	if msg.From != "" {
		from = msg.From
	}

	ok, _, _, _, resp := p.client.External(
		"POST",
		baseurl+"/api/sms/send",
		map[string]any{
			"to":      msg.To,
			"from":    from,
			"sms":     msg.Text,
			"type":    "plain",
			"channel": "generic",
			"api_key": apiKey,
		},
		nil,
		map[string]string{
			"accept":       "application/json",
			"content-type": "application/json",
		},
		0,
	)

	body, _ := resp.(map[string]any)
	return ok && body["code"] == "ok"
}

func (p *TermiiProvider) Mobile(_ MobileMessage) bool { return false }
