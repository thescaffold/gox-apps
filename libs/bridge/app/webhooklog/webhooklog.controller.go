package webhooklog

import (
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
)

// WebhookLogController mirrors ntx-apps/libs/bridge/src/api/webhook-log/webhook-log.controller.ts.
type WebhookLogController struct {
	crud.CrudResource[WebhookLog, CreateWebhookLogDto, UpdateWebhookLogDto]

	entity *WebhookLogEntity `inject:""`
	lang   *i18n.Service     `inject:""`
}

func (c *WebhookLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[WebhookLog, CreateWebhookLogDto, UpdateWebhookLogDto]{
		// Mirrors TS webhook-log.controller.ts:18 name = 'webhook log'.
		Name: "webhook log",
		// Mirrors TS webhook-log.controller.ts:19 searchable = [] (empty list).
		Searchable: []string{},
		// TS unique = () => [] — no uniqueness constraints; left unset.
	})
	c.SetLang(c.lang)
}
