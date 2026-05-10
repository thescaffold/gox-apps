package eventlog

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type EventLog struct {
	sql.BaseEntity
	EventId  string          `gorm:"column:event_id;type:varchar(36);not null" json:"eventId"`
	Meta     json.RawMessage `gorm:"column:meta;type:jsonb"                    json:"meta,omitempty"`
	Request  json.RawMessage `gorm:"column:request;type:jsonb"                 json:"request,omitempty"`
	Response json.RawMessage `gorm:"column:response;type:jsonb"                json:"response,omitempty"`
	Status   *string         `gorm:"column:status;type:varchar(255)"           json:"status,omitempty"`
}

func (EventLog) TableName() string { return "PolylogEventLogs" }

type EventLogEntity struct {
	*sql.Entity[EventLog] `inject:""`
}

func (e *EventLogEntity) OnRegister() {
	e.Hydrate("PolylogEventLogs", []string{"event_id"}, nil, nil, nil, nil, nil, "created_at desc")
}
