package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	ctrlreq "github.com/thescaffold/gox-apps/libs/controller/app/request"
	ctrlroute "github.com/thescaffold/gox-apps/libs/controller/app/route"
	"github.com/thescaffold/gox-apps/libs/controller/migrations"
	"github.com/thescaffold/gox-packages/libs/core/module"
)

var Migrations = []sql.Migration{
	&migrations.CreateControllerRoutes{},
	&migrations.CreateControllerRequests{},
}

type AppModule struct{}

func (m *AppModule) Imports() []types.Module {
	return []types.Module{
		module.New(module.CoreConfig{}),
		sql.Child(&sql.Config{Migrations: Migrations}),
		&ctrlroute.RouteModule{},
		&ctrlreq.RequestModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	return []any{&AppController{}, &AppService{}}
}
