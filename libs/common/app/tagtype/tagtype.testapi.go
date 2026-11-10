package tagtype

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/thescaffold/gox-apps/libs/common/app/tag"
	"gorm.io/gorm"
)

// NewControllerForTest wires a TagTypeController over db without the DI container, for
// integration tests of sync/attach/detach.
func NewControllerForTest(db *gorm.DB) *TagTypeController {
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	te := &TagTypeEntity{Entity: (&sql.Entity[TagType]{}).With(q)}
	te.OnRegister()
	tg := &tag.TagEntity{Entity: (&sql.Entity[tag.Tag]{}).With(q)}
	tg.OnRegister()
	return &TagTypeController{entity: te, service: &TagTypeService{entity: te}, tagService: tag.NewTagServiceForTest(tg)}
}
