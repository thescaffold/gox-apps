package rule

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

type Rule struct {
	sql.BaseEntity

	Group  string          `gorm:"column:group;type:varchar(255);not null" json:"group"`
	Key    string          `gorm:"column:key;type:varchar(255);not null"   json:"key"`
	Rules  json.RawMessage `gorm:"column:rules;type:jsonb;not null"         json:"rules"`
	Status *string         `gorm:"column:status;type:varchar(255)"          json:"status,omitempty"`
}

func (Rule) TableName() string { return "NotificationRules" }

type RuleEntity struct {
	*sql.Entity[Rule] `inject:""`
}

func (e *RuleEntity) OnRegister() {
	e.Hydrate("NotificationRules", []string{"group", "key"}, nil, nil, nil, nil, nil, "created_at desc")
}
