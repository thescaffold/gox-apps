package webhook

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
)

type WebhookModule struct{}

func (m *WebhookModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}
func (m *WebhookModule) Exports() []any      { return []any{&WebhookService{}} }
func (m *WebhookModule) Declarations() []any { return []any{&WebhookService{}, &WebhookEntity{}} }
