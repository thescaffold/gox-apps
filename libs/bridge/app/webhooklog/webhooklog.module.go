package webhooklog
import ("github.com/awesome-goose/goose/modules/sql"; "github.com/awesome-goose/goose/types")
type WebhookLogModule struct{}
func (m *WebhookLogModule) Imports() []types.Module { return []types.Module{ROUTES, sql.Child(&sql.Config{})} }
func (m *WebhookLogModule) Exports() []any { return []any{&WebhookLogService{}} }
func (m *WebhookLogModule) Declarations() []any { return []any{&WebhookLogService{}, &WebhookLogEntity{}} }
