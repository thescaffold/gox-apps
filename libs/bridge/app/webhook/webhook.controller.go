package webhook
import "github.com/thescaffold/gox-packages-core/crud"
type WebhookController struct {
	crud.CrudResource[Webhook, CreateWebhookDto, UpdateWebhookDto]
	entity *WebhookEntity `inject:""`
}
func (c *WebhookController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Webhook, CreateWebhookDto, UpdateWebhookDto]{
		Name: "BridgeWebhook", Searchable: []string{"license_id", "type"},
	})
}
