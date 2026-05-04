package list

import (
	"encoding/json"
	"time"

	"github.com/awesome-goose/goose/modules/sql"
)

type List struct {
	sql.BaseEntity

	Key       string          `gorm:"column:key;type:varchar(255);not null"    json:"key"`
	Value     json.RawMessage `gorm:"column:value;type:jsonb"                  json:"value,omitempty"`
	Group     string          `gorm:"column:group;type:varchar(255);not null"  json:"group"`
	ExpiredAt *time.Time      `gorm:"column:expired_at"                        json:"expiredAt,omitempty"`
	Meta      json.RawMessage `gorm:"column:meta;type:jsonb"                   json:"meta,omitempty"`
	Status    *string         `gorm:"column:status;type:varchar(255)"          json:"status,omitempty"`
}

func (List) TableName() string { return "CacheLists" }

type ListEntity struct {
	*sql.Entity[List] `inject:""`
}

func (e *ListEntity) OnRegister() {
	e.Hydrate(
		"CacheLists",
		[]string{"key", "group"},
		[]string{},
		nil,
		func(l *List) (any, []any) {
			return map[string]any{"key": l.Key, "group": l.Group}, nil
		},
		nil,
		nil,
		"created_at desc",
	)
}
