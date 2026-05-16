package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	flagsenvironment "github.com/thescaffold/gox-apps/libs/flags/app/environment"
	flagsenvironmenttype "github.com/thescaffold/gox-apps/libs/flags/app/environmenttype"
	flagsflag "github.com/thescaffold/gox-apps/libs/flags/app/flag"
	flagsflaglog "github.com/thescaffold/gox-apps/libs/flags/app/flaglog"
	"github.com/thescaffold/gox-apps/libs/flags/migrations"
	"github.com/thescaffold/gox-packages/libs/core/module"
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
		// TranslationPaths mirror TS index.ts `paths = [translations/en/ntx/apps/${name}.yaml]`;
		// CoreModule.Boot pre-loads them so translate() resolves real strings.
		module.New(module.CoreConfig{
			TranslationPaths: Paths,
		}),
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
	svc := &AppService{}
	flagsAppSvc = svc
	return []any{&AppController{}, svc}
}
