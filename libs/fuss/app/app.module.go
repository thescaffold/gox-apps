package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	fusshist "github.com/thescaffold/gox-apps-fuss/app/history"
	fusstok "github.com/thescaffold/gox-apps-fuss/app/token"
	fusstoklog "github.com/thescaffold/gox-apps-fuss/app/tokenlog"
	"github.com/thescaffold/gox-apps-fuss/migrations"
	"github.com/thescaffold/gox-packages-core/module"
)

var Migrations = []sql.Migration{
	&migrations.CreateFussTokens{},
	&migrations.CreateFussTokenLogs{},
	&migrations.CreateFussHistories{},
}

type AppModule struct{}

func (m *AppModule) Imports() []types.Module {
	return []types.Module{
		module.New(module.CoreConfig{}),
		sql.Child(&sql.Config{Migrations: Migrations}),
		&fusstok.TokenModule{},
		&fusstoklog.TokenLogModule{},
		&fusshist.HistoryModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	svc := &AppService{}
	fussAppSvc = svc
	return []any{&AppController{}, svc}
}
