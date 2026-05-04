package workspace

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

type Workspace struct {
	sql.BaseEntity
	UserId    string          `gorm:"column:user_id;type:varchar(36);not null"   json:"userId"`
	ClientId  string          `gorm:"column:client_id;type:varchar(255);not null" json:"clientId"`
	Name      string          `gorm:"column:name;type:varchar(255);not null"      json:"name"`
	Desc      *string         `gorm:"column:desc;type:varchar(255)"               json:"desc,omitempty"`
	Slug      string          `gorm:"column:slug;type:varchar(255);not null"      json:"slug"`
	LogoUrl   *string         `gorm:"column:logo_url;type:varchar(255)"           json:"logoUrl,omitempty"`
	BannerUrl *string         `gorm:"column:banner_url;type:varchar(255)"         json:"bannerUrl,omitempty"`
	Meta      json.RawMessage `gorm:"column:meta;type:jsonb"                      json:"meta,omitempty"`
	Status    *string         `gorm:"column:status;type:varchar(255)"             json:"status,omitempty"`
}
func (Workspace) TableName() string { return "IdentityWorkspaces" }
type WorkspaceEntity struct{ *sql.Entity[Workspace] `inject:""` }
func (e *WorkspaceEntity) OnRegister() {
	e.Hydrate("IdentityWorkspaces", []string{"user_id", "client_id", "slug"}, nil, nil, nil, nil, nil, "created_at desc")
}
