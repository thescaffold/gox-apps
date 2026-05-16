package formtype

import (
	"encoding/json"

	"github.com/awesome-goose/goose/modules/sql"
)

// FormType mirrors ntx-apps/libs/forms/src/api/form-type/entities/form-type.entity.ts.
type FormType struct {
	sql.BaseEntity

	Name         string  `gorm:"column:name;type:varchar(255);not null"  json:"name"`
	Desc         *string `gorm:"column:desc;type:varchar(255)"           json:"desc,omitempty"`
	Detail       *string `gorm:"column:detail;type:text"                 json:"detail,omitempty"`
	ThumbnailUrl *string `gorm:"column:thumbnail_url;type:varchar(255)"  json:"thumbnailUrl,omitempty"`
	BannerUrl    *string `gorm:"column:banner_url;type:varchar(255)"     json:"bannerUrl,omitempty"`
	// Tags + Meta are jsonb blobs — TS exposes them as `string[]` and any
	// respectively, so they must round-trip as raw JSON.
	Tags json.RawMessage `gorm:"column:tags;type:jsonb" json:"tags,omitempty"`
	Meta json.RawMessage `gorm:"column:meta;type:jsonb" json:"meta,omitempty"`
}

func (FormType) TableName() string { return "FormFormTypes" }

type FormTypeEntity struct {
	*sql.Entity[FormType] `inject:""`
}

func (e *FormTypeEntity) OnRegister() {
	// Searchable mirrors TS form-type.controller.ts:19 searchable = ['name','desc','tags'].
	e.Hydrate("FormFormTypes", []string{"name", "desc", "tags"}, nil, nil, nil, nil, nil, "created_at desc")
}
