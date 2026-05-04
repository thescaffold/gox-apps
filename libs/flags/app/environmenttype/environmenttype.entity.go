package environmenttype

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

type EnvironmentType struct {
	sql.BaseEntity

	Category     string          `gorm:"column:category;type:varchar(255);not null"  json:"category"`
	Name         string          `gorm:"column:name;type:varchar(255);not null"      json:"name"`
	Desc         *string         `gorm:"column:desc;type:varchar(255)"               json:"desc,omitempty"`
	Detail       *string         `gorm:"column:detail;type:text"                     json:"detail,omitempty"`
	ThumbnailUrl *string         `gorm:"column:thumbnail_url;type:varchar(255)"      json:"thumbnailUrl,omitempty"`
	BannerUrl    *string         `gorm:"column:banner_url;type:varchar(255)"         json:"bannerUrl,omitempty"`
	Tags         json.RawMessage `gorm:"column:tags;type:jsonb"                      json:"tags,omitempty"`
	Meta         json.RawMessage `gorm:"column:meta;type:jsonb"                      json:"meta,omitempty"`
}

func (EnvironmentType) TableName() string { return "FlagEnvironmentTypes" }

type EnvironmentTypeEntity struct {
	*sql.Entity[EnvironmentType] `inject:""`
}

func (e *EnvironmentTypeEntity) OnRegister() {
	e.Hydrate("FlagEnvironmentTypes", []string{"category", "name"}, nil, nil, nil, nil, nil, "created_at desc")
}
