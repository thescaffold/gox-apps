package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	polylogconfig "github.com/thescaffold/gox-apps/libs/polylog/app/config"
	polylogsource "github.com/thescaffold/gox-apps/libs/polylog/app/source"
	polylogsourcetype "github.com/thescaffold/gox-apps/libs/polylog/app/sourcetype"
	"github.com/thescaffold/gox-apps/libs/polylog/migrations"
	polylogpipeline "github.com/thescaffold/gox-apps/libs/polylog/pkg/pipeline"
	"github.com/thescaffold/gox-packages/libs/core/module"
)

var Migrations = []sql.Migration{
	&migrations.CreatePolylogSourceTypes{},
	&migrations.CreatePolylogSources{},
	&migrations.CreatePolylogSinkTypes{},
	&migrations.CreatePolylogSinks{},
	&migrations.CreatePolylogChannels{},
	&migrations.CreatePolylogEvents{},
	&migrations.CreatePolylogEventLogs{},
	&migrations.CreatePolylogConfigs{},
}

type AppModule struct{}

func (m *AppModule) Imports() []types.Module {
	return []types.Module{
		module.New(module.CoreConfig{}),
		sql.Child(&sql.Config{Migrations: Migrations}),
		&polylogsourcetype.SourceTypeModule{},
		&polylogsource.SourceModule{},
		&polylogconfig.ConfigModule{},
		&polylogpipeline.PipelineModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	return []any{&AppController{}, &AppService{}}
}
