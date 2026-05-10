package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/blobs/app/file"
	"github.com/thescaffold/gox-apps/libs/blobs/app/page"
	"github.com/thescaffold/gox-apps/libs/blobs/migrations"
	"github.com/thescaffold/gox-packages/libs/core/module"
)

var Migrations = []sql.Migration{
	&migrations.CreateBlobsFiles{},
	&migrations.CreateBlobsPages{},
}

type AppModule struct{}

func (m *AppModule) Imports() []types.Module {
	return []types.Module{
		module.New(module.CoreConfig{}),
		sql.Child(&sql.Config{
			Migrations: Migrations,
		}),
		&file.FileModule{},
		&page.PageModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any {
	return []any{&AppService{}}
}

func (m *AppModule) Declarations() []any {
	return []any{&AppService{}}
}
