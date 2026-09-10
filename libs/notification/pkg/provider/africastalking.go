package provider

import (
	"os"

	ntxhttp "github.com/thescaffold/gox-packages/libs/core/http"
)

type AfricastalkingProvider struct {
	client *ntxhttp.Client `inject:""`
}

func (p *AfricastalkingProvider) Email(_ EmailMessage) bool { return false }

func (p *AfricastalkingProvider) SMS(msg SMSMessage) bool {
	baseurl := os.Getenv("AFRICASTALKING_BASE_URL")
	username := os.Getenv("AFRICASTALKING_USERNAME")
	apiKey := os.Getenv("AFRICASTALKING_API_KEY")
	senderID := os.Getenv("AFRICASTALKING_SENDER_ID")

	from := senderID
	if msg.From != "" {
		from = msg.From
	}

	// Africa's Talking bulk SMS: POST /messaging/bulk (JSON), apiKey header. A
	// 2xx carries {"SMSMessageData": {"Recipients": [{"status": "Success", ...}]}}.
	ok, _, _, _, resp := p.client.External(
		"POST",
		baseurl+"/messaging/bulk",
		map[string]any{
			"username":     username,
			"message":      msg.Text,
			"senderId":     from,
			"phoneNumbers": []string{msg.To},
			"bulkSMSMode":  1,
		},
		nil,
		map[string]string{
			"apiKey":       apiKey,
			"accept":       "application/json",
			"content-type": "application/json",
		},
		0,
	)

	if !ok {
		return false
	}

	body, _ := resp.(map[string]any)
	data, _ := body["SMSMessageData"].(map[string]any)
	recipients, _ := data["Recipients"].([]any)
	for _, r := range recipients {
		if rec, ok := r.(map[string]any); ok && rec["status"] == "Success" {
			return true
		}
	}
	return false
}

func (p *AfricastalkingProvider) Mobile(_ MobileMessage) bool { return false }
