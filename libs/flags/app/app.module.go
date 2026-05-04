package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	flagsenvironment "github.com/thescaffold/gox-apps-flags/app/environment"
	flagsenvironmenttype "github.com/thescaffold/gox-apps-flags/app/environmenttype"
	flagsflag "github.com/thescaffold/gox-apps-flags/app/flag"
	flagsflaglog "github.com/thescaffold/gox-apps-flags/app/flaglog"
	"github.com/thescaffold/gox-apps-flags/migrations"
	"github.com/thescaffold/gox-packages-core/module"
)

var Migrations = []sql.Migration{
	&migrations.CreateFlagEnvironmentTypes{},
	&migrations.CreateFlagEnvironments{},
	&migrations.CreateFlagFlags{},
	&migrations.CreateFlagFlagLogs{},
}

type AppModule struct{}

func (m *AppModule) Imports() []types.Module {
	return []types.Module{
		module.New(module.CoreConfig{}),
		sql.Child(&sql.Config{Migrations: Migrations}),
		&flagsenvironmenttype.EnvironmentTypeModule{},
		&flagsenvironment.EnvironmentModule{},
		&flagsflag.FlagModule{},
		&flagsflaglog.FlagLogModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	return []any{&AppController{}, &AppService{}}
}
