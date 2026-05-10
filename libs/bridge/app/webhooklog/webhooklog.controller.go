package webhooklog

import "github.com/thescaffold/gox-packages-core/crud"

type WebhookLogController struct {
	crud.CrudResource[WebhookLog, CreateWebhookLogDto, UpdateWebhookLogDto]
	entity *WebhookLogEntity `inject:""`
}

func (c *WebhookLogController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[WebhookLog, CreateWebhookLogDto, UpdateWebhookLogDto]{
		Name: "BridgeWebhookLog", Searchable: []string{"webhook_id"},
	})
}
