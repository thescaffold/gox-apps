package formtype

import "github.com/awesome-goose/goose/modules/sql"

type FormType struct {
	sql.BaseEntity

	Name         string  `gorm:"column:name;type:varchar(255);not null"  json:"name"`
	Desc         *string `gorm:"column:desc;type:varchar(255)"           json:"desc,omitempty"`
	Detail       *string `gorm:"column:detail;type:text"                 json:"detail,omitempty"`
	ThumbnailUrl *string `gorm:"column:thumbnail_url;type:varchar(255)"  json:"thumbnailUrl,omitempty"`
	BannerUrl    *string `gorm:"column:banner_url;type:varchar(255)"     json:"bannerUrl,omitempty"`
	Tags         *string `gorm:"column:tags;type:jsonb"                  json:"tags,omitempty"`
	Meta         *string `gorm:"column:meta;type:jsonb"                  json:"meta,omitempty"`
}

func (FormType) TableName() string { return "FormFormTypes" }

type FormTypeEntity struct {
	*sql.Entity[FormType] `inject:""`
}

func (e *FormTypeEntity) OnRegister() {
	e.Hydrate("FormFormTypes", []string{"name"}, nil, nil, nil, nil, nil, "created_at desc")
}
