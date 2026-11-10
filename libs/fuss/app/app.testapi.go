package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/thescaffold/gox-apps/libs/fuss/app/history"
	"github.com/thescaffold/gox-apps/libs/fuss/app/token"
	"github.com/thescaffold/gox-apps/libs/fuss/app/tokenlog"
	"gorm.io/gorm"
)

// NewAppServiceForTest builds an AppService over db without DI.
func NewAppServiceForTest(db *gorm.DB) *AppService {
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	te := &token.TokenEntity{Entity: (&sql.Entity[token.Token]{}).With(q)}
	te.OnRegister()
	le := &tokenlog.TokenLogEntity{Entity: (&sql.Entity[tokenlog.TokenLog]{}).With(q)}
	le.OnRegister()
	he := &history.HistoryEntity{Entity: (&sql.Entity[history.History]{}).With(q)}
	he.OnRegister()
	return &AppService{tokenEntity: te, tokenLogEntity: le, historyEntity: he}
}
