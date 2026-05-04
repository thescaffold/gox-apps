package webhooklog
import "encoding/json"
type CreateWebhookLogDto struct {
	WebhookId string          `json:"webhookId" binding:"required"`
	Type      *string         `json:"type,omitempty"`
	Meta      json.RawMessage `json:"meta,omitempty"`
	Request   json.RawMessage `json:"request,omitempty"`
	Response  json.RawMessage `json:"response,omitempty"`
	Status    *string         `json:"status,omitempty"`
}
type UpdateWebhookLogDto struct{ Status *string `json:"status,omitempty"` }
