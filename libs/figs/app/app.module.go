package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps-figs/app/file"
	"github.com/thescaffold/gox-apps-figs/migrations"
	"github.com/thescaffold/gox-apps-figs/pkg/converter"
	"github.com/thescaffold/gox-apps-figs/pkg/mapper"
	"github.com/thescaffold/gox-apps-figs/pkg/store"
	"github.com/thescaffold/gox-apps-figs/pkg/validator"
	"github.com/thescaffold/gox-packages-core/module"
)

var Migrations = []sql.Migration{
	&migrations.CreateFigsFiles{},
}

type AppModule struct{}

func (m *AppModule) Imports() []types.Module {
	return []types.Module{
		module.New(module.CoreConfig{}),
		sql.Child(&sql.Config{
			Migrations: Migrations,
		}),
		&file.FileModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any {
	return []any{&AppService{}}
}

func (m *AppModule) Declarations() []any {
	svc := &AppService{}
	figsAppSvc = svc
	return []any{
		&AppController{},
		svc,
		&converter.Service{},
		&mapper.Service{},
		&store.Service{},
		&validator.Service{},
	}
}
