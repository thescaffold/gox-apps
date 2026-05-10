package provider

import (
	"github.com/awesome-goose/goose/types"
	notificationlog "github.com/thescaffold/gox-apps/libs/notification/app/log"
	notificationmessage "github.com/thescaffold/gox-apps/libs/notification/app/message"
	notificationrule "github.com/thescaffold/gox-apps/libs/notification/app/rule"
	notificationtemplate "github.com/thescaffold/gox-apps/libs/notification/app/template"
)

type ProviderModule struct{}

func (m *ProviderModule) Imports() []types.Module {
	return []types.Module{
		&notificationtemplate.TemplateModule{},
		&notificationrule.RuleModule{},
		&notificationlog.LogModule{},
		&notificationmessage.MessageModule{},
	}
}

func (m *ProviderModule) Exports() []any { return []any{&ProviderService{}} }

func (m *ProviderModule) Declarations() []any {
	return []any{
		&ProviderService{},
		&MailgunProvider{},
		&ZohoProvider{},
	}
}
