package request

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

type Request struct {
	sql.BaseEntity

	Group    *string         `gorm:"column:group;type:varchar(255)"             json:"group,omitempty"`
	Service  *string         `gorm:"column:service;type:varchar(255)"           json:"service,omitempty"`
	Type     string          `gorm:"column:type;type:varchar(255);default:'http'" json:"type"`
	Method   string          `gorm:"column:method;type:varchar(255);not null"   json:"method"`
	Url      string          `gorm:"column:url;type:varchar(255);not null"      json:"url"`
	Queries  json.RawMessage `gorm:"column:queries;type:jsonb"                  json:"queries,omitempty"`
	Body     json.RawMessage `gorm:"column:body;type:jsonb"                     json:"body,omitempty"`
	Response json.RawMessage `gorm:"column:response;type:jsonb"                 json:"response,omitempty"`
	Status   *string         `gorm:"column:status;type:varchar(255)"            json:"status,omitempty"`
}

func (Request) TableName() string { return "ControllerRequests" }

type RequestEntity struct {
	*sql.Entity[Request] `inject:""`
}

func (e *RequestEntity) OnRegister() {
	e.Hydrate("ControllerRequests", []string{"group", "service", "method", "url", "type"}, nil, nil, nil, nil, nil, "created_at desc")
}
