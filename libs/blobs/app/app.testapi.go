package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/thescaffold/gox-apps/libs/blobs/app/file"
	"github.com/thescaffold/gox-apps/libs/blobs/app/page"
	"gorm.io/gorm"
)

// NewAppServiceForTest builds an AppService over db without DI.
func NewAppServiceForTest(db *gorm.DB) *AppService {
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	fe := &file.FileEntity{Entity: (&sql.Entity[file.File]{}).With(q)}
	fe.OnRegister()
	pe := &page.PageEntity{Entity: (&sql.Entity[page.Page]{}).With(q)}
	pe.OnRegister()
	return &AppService{fileEntity: fe, pageEntity: pe}
}
