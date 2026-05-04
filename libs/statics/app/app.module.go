package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps-statics/app/list"
	"github.com/thescaffold/gox-apps-statics/migrations"
	"github.com/thescaffold/gox-packages-core/module"
)

// Migrations contains schema migrations run by this app.
var Migrations = []sql.Migration{
	&migrations.CreateStaticsLists{},
}

// Seeders contains data seeders run after migrations.
var Seeders = []sql.Seeder{
	&migrations.SeedStaticsListsData{},
}

type AppModule struct{}

func (m *AppModule) Imports() []types.Module {
	return []types.Module{
		module.New(module.CoreConfig{}),
		sql.Child(&sql.Config{
			Migrations: Migrations,
			Seeders:    Seeders,
		}),
		&list.ListModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any {
	return []any{&AppService{}}
}

func (m *AppModule) Declarations() []any {
	return []any{
		&AppService{},
	}
}
