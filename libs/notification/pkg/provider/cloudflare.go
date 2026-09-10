package provider

import (
	"fmt"
	"os"

	ntxhttp "github.com/thescaffold/gox-packages/libs/core/http"
)

type CloudflareProvider struct {
	client *ntxhttp.Client `inject:""`
}

func (p *CloudflareProvider) Email(msg EmailMessage) bool {
	baseurl := os.Getenv("CLOUDFLARE_BASE_URL")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	token := os.Getenv("CLOUDFLARE_API_TOKEN")
	fromEmail := os.Getenv("CLOUDFLARE_FROM_EMAIL")
	fromName := os.Getenv("CLOUDFLARE_FROM_NAME")

	from := fromEmail
	if msg.From != "" {
		from = msg.From
	}
	if fromName != "" {
		from = fmt.Sprintf("%s <%s>", fromName, from)
	}

	// Cloudflare Email Service REST API:
	// POST /accounts/{account_id}/email/sending/send, Bearer auth. A 200 carries
	// {"success": true, "result": {"delivered": [...], "queued": [...]}}.
	ok, _, _, _, resp := p.client.External(
		"POST",
		fmt.Sprintf("%s/accounts/%s/email/sending/send", baseurl, accountID),
		map[string]any{
			"from":    from,
			"to":      msg.To,
			"subject": msg.Subject,
			"text":    msg.Text,
			"html":    msg.HTML,
		},
		nil,
		map[string]string{
			"authorization": "Bearer " + token,
			"accept":        "application/json",
			"content-type":  "application/json",
		},
		0,
	)

	body, _ := resp.(map[string]any)
	return ok && body["success"] == true
}

func (p *CloudflareProvider) SMS(_ SMSMessage) bool       { return false }
func (p *CloudflareProvider) Mobile(_ MobileMessage) bool { return false }
