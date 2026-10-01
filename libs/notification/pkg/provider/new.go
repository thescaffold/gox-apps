package provider

import (
	"github.com/awesome-goose/goose/modules/sql"
	notificationlog "github.com/thescaffold/gox-apps/libs/notification/app/log"
	notificationmessage "github.com/thescaffold/gox-apps/libs/notification/app/message"
	notificationrule "github.com/thescaffold/gox-apps/libs/notification/app/rule"
	notificationtemplate "github.com/thescaffold/gox-apps/libs/notification/app/template"
	ntxhttp "github.com/thescaffold/gox-packages/libs/core/http"
	"gorm.io/gorm"
)

// NewService builds a ProviderService over db without the DI container, for a
// process that composes its own services (and for tests). hmacKey signs the
// outgoing provider requests.
func NewService(db *gorm.DB, hmacKey string) *ProviderService {
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	client := ntxhttp.New(hmacKey)

	tpl := &notificationtemplate.TemplateEntity{Entity: (&sql.Entity[notificationtemplate.Template]{}).With(q)}
	rule := &notificationrule.RuleEntity{Entity: (&sql.Entity[notificationrule.Rule]{}).With(q)}
	logs := &notificationlog.LogEntity{Entity: (&sql.Entity[notificationlog.Log]{}).With(q)}
	msgs := &notificationmessage.MessageEntity{Entity: (&sql.Entity[notificationmessage.Message]{}).With(q)}
	for _, e := range []interface{ OnRegister() }{tpl, rule, logs, msgs} {
		e.OnRegister()
	}
	return &ProviderService{
		templateEntity: tpl, ruleEntity: rule, logEntity: logs, messageEntity: msgs,
		mailgun: &MailgunProvider{client: client}, zoho: &ZohoProvider{client: client},
		cloudflare: &CloudflareProvider{client: client}, termii: &TermiiProvider{client: client},
		africastalking: &AfricastalkingProvider{client: client},
	}
}
