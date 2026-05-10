package provider

import (
	"encoding/base64"
	"fmt"
	"os"

	ntxhttp "github.com/thescaffold/gox-packages-core/http"
)

type MailgunProvider struct {
	client *ntxhttp.Client `inject:""`
}

func (p *MailgunProvider) Email(msg EmailMessage) bool {
	baseurl := os.Getenv("MAILGUN_BASE_URL")
	domain := os.Getenv("MAILGUN_DOMAIN")
	token := os.Getenv("MAILGUN_TOKEN")
	from := os.Getenv("MAILGUN_FROM")
	if msg.From != "" {
		from = msg.From
	}

	creds := base64.StdEncoding.EncodeToString([]byte("api:" + token))
	ok, _, _, _, _ := p.client.External(
		"POST",
		fmt.Sprintf("%s/v3/%s/messages", baseurl, domain),
		map[string]any{
			"from":    from,
			"to":      msg.To,
			"subject": msg.Subject,
			"text":    msg.Text,
			"html":    msg.HTML,
		},
		nil,
		map[string]string{
			"authorization": "basic " + creds,
			"content-type":  "multipart/form-data",
		},
		0,
	)
	return ok
}

func (p *MailgunProvider) SMS(_ SMSMessage) bool       { return false }
func (p *MailgunProvider) Mobile(_ MobileMessage) bool { return false }
