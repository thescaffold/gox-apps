package providerlog

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type ProviderLog struct {
	sql.BaseEntity
	ProviderId string          `gorm:"column:provider_id;type:varchar(36);not null" json:"providerId"`
	Event      string          `gorm:"column:event;type:varchar(255);not null"      json:"event"`
	Request    json.RawMessage `gorm:"column:request;type:jsonb"                    json:"request,omitempty"`
	Response   json.RawMessage `gorm:"column:response;type:jsonb"                   json:"response,omitempty"`
	Meta       json.RawMessage `gorm:"column:meta;type:jsonb"                       json:"meta,omitempty"`
	Status     *string         `gorm:"column:status;type:varchar(255)"              json:"status,omitempty"`
}
func (ProviderLog) TableName() string { return "IdentityProviderLogs" }
type ProviderLogEntity struct{ *sql.Entity[ProviderLog] `inject:""` }
func (e *ProviderLogEntity) OnRegister() {
	e.Hydrate("IdentityProviderLogs", []string{"provider_id", "event"}, nil, nil, nil, nil, nil, "created_at desc")
}
