package sinktype

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type SinkType struct {
	sql.BaseEntity
	Category     string          `gorm:"column:category;type:varchar(255);not null"  json:"category"`
	Key          *string         `gorm:"column:key;type:varchar(255)"                json:"key,omitempty"`
	Visibility   *string         `gorm:"column:visibility;type:varchar(255)"         json:"visibility,omitempty"`
	Name         string          `gorm:"column:name;type:varchar(255);not null"      json:"name"`
	Desc         *string         `gorm:"column:desc;type:varchar(255)"               json:"desc,omitempty"`
	Detail       *string         `gorm:"column:detail;type:text"                     json:"detail,omitempty"`
	ThumbnailUrl *string         `gorm:"column:thumbnail_url;type:varchar(255)"      json:"thumbnailUrl,omitempty"`
	BannerUrl    *string         `gorm:"column:banner_url;type:varchar(255)"         json:"bannerUrl,omitempty"`
	Tags         json.RawMessage `gorm:"column:tags;type:jsonb"                      json:"tags,omitempty"`
	Meta         json.RawMessage `gorm:"column:meta;type:jsonb"                      json:"meta,omitempty"`
}
func (SinkType) TableName() string { return "PolylogSinkTypes" }
type SinkTypeEntity struct{ *sql.Entity[SinkType] `inject:""` }
func (e *SinkTypeEntity) OnRegister() {
	e.Hydrate("PolylogSinkTypes", []string{"category", "name"}, nil, nil, nil, nil, nil, "created_at desc")
}
