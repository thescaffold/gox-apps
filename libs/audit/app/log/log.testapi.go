package log

import (
	"github.com/awesome-goose/goose/modules/sql"
	"gorm.io/gorm"
)

// NewLogEntityForTest and NewLogServiceForTest construct a LogEntity/
// LogService wired to the given gorm.DB, bypassing the DI container —
// mirrors the ForTest pattern used across Origine's own gox-apps (e.g.
// gox-apps/libs/systems/app/project.testapi.go), added here so
// PLAN M1-04's actor fields can be verified against a real Postgres row
// through the actual Activities() query, not just a struct literal.
func NewLogEntityForTest(db *gorm.DB) *LogEntity {
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	e := (&sql.Entity[Log]{}).With(q)
	entity := &LogEntity{Entity: e}
	entity.OnRegister()
	return entity
}

func NewLogServiceForTest(db *gorm.DB) *LogService {
	return &LogService{entity: NewLogEntityForTest(db)}
}
