package webhook

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// WebhookController mirrors ntx-apps/libs/bridge/src/api/webhook/webhook.controller.ts.
type WebhookController struct {
	crud.CrudResource[Webhook, CreateWebhookDto, UpdateWebhookDto]

	entity *WebhookEntity `inject:""`
	lang   *i18n.Service  `inject:""`
}

func (c *WebhookController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Webhook, CreateWebhookDto, UpdateWebhookDto]{
		// Mirrors TS webhook.controller.ts:18 name = 'webhook'.
		Name: "webhook",
		// Mirrors TS webhook.controller.ts:19 searchable = [] (empty list).
		Searchable: []string{},
		// Mirrors TS webhook.controller.ts:21-23 unique = ({licenseId,type,url})
		// => [{licenseId,type,url}].
		Unique: func(d *CreateWebhookDto) []map[string]any {
			return []map[string]any{{
				"license_id": d.LicenseId,
				"type":       d.Type,
				"url":        d.Url,
			}}
		},
	})
	c.SetLang(c.lang)
}
