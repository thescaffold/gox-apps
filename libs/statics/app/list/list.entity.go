package list

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

type ListKeyType = string

const (
	ListKeyTypeSetting  ListKeyType = "setting"
	ListKeyTypeCountry  ListKeyType = "country"
	ListKeyTypeCity     ListKeyType = "city"
	ListKeyTypeCurrency ListKeyType = "currency"
	ListKeyTypeLanguage ListKeyType = "language"
	ListKeyTypeTimezone ListKeyType = "timezone"
	ListKeyTypeGender   ListKeyType = "gender"
	ListKeyTypeIndustry ListKeyType = "industry"
)

type List struct {
	sql.BaseEntity

	ParentId *string         `gorm:"column:parent_id;type:varchar(36)"          json:"parentId,omitempty"`
	Key      string          `gorm:"column:key;type:varchar(255);not null"      json:"key"`
	Code     *string         `gorm:"column:code;type:varchar(255)"              json:"code,omitempty"`
	Value    string          `gorm:"column:value;type:varchar(255);not null"    json:"value"`
	Meta     json.RawMessage `gorm:"column:meta;type:jsonb"                     json:"meta,omitempty"`
	Status   *string         `gorm:"column:status;type:varchar(255)"            json:"status,omitempty"`
}

func (List) TableName() string { return "StaticsLists" }

type ListEntity struct {
	*sql.Entity[List] `inject:""`
}

func (e *ListEntity) OnRegister() {
	e.Hydrate(
		"StaticsLists",
		[]string{"value", "code", "key"},
		[]string{},
		nil,
		func(l *List) (any, []any) {
			return map[string]any{"parent_id": l.ParentId, "key": l.Key}, nil
		},
		nil,
		nil,
		"created_at desc",
	)
}
