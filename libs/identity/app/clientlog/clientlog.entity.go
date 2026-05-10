package clientlog

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type ClientLog struct {
	sql.BaseEntity
	ClientId string          `gorm:"column:client_id;type:varchar(255);not null" json:"clientId"`
	Event    string          `gorm:"column:event;type:varchar(255);not null"     json:"event"`
	Request  json.RawMessage `gorm:"column:request;type:jsonb"                   json:"request,omitempty"`
	Response json.RawMessage `gorm:"column:response;type:jsonb"                  json:"response,omitempty"`
	Meta     json.RawMessage `gorm:"column:meta;type:jsonb"                      json:"meta,omitempty"`
	Status   *string         `gorm:"column:status;type:varchar(255)"             json:"status,omitempty"`
}

func (ClientLog) TableName() string { return "IdentityClientLogs" }

type ClientLogEntity struct {
	*sql.Entity[ClientLog] `inject:""`
}

func (e *ClientLogEntity) OnRegister() {
	e.Hydrate("IdentityClientLogs", []string{"client_id", "event"}, nil, nil, nil, nil, nil, "created_at desc")
}
