package provider

import (
	"os"

	ntxhttp "github.com/thescaffold/gox-packages/libs/core/http"
)

type ZohoProvider struct {
	client *ntxhttp.Client `inject:""`
}

func (p *ZohoProvider) Email(msg EmailMessage) bool {
	baseurl := os.Getenv("ZOHO_BASE_URL")
	token := os.Getenv("ZOHO_TOKEN")
	fromEmail := os.Getenv("ZOHO_FROM_EMAIL")
	fromName := os.Getenv("ZOHO_FROM_NAME")

	from := fromEmail
	if msg.From != "" {
		from = msg.From
	}

	ok, _, _, _, _ := p.client.External(
		"POST",
		baseurl+"/v1.1/email",
		map[string]any{
			"from": map[string]any{
				"address": from,
				"name":    fromName,
			},
			"to": []any{
				map[string]any{
					"email_address": map[string]any{"address": msg.To},
				},
			},
			"subject":  msg.Subject,
			"htmlbody": msg.HTML,
		},
		nil,
		map[string]string{
			"authorization": "Zoho-enczapikey " + token,
			"accept":        "application/json",
			"content-type":  "application/json",
		},
		0,
	)
	return ok
}

func (p *ZohoProvider) SMS(_ SMSMessage) bool       { return false }
func (p *ZohoProvider) Mobile(_ MobileMessage) bool { return false }
