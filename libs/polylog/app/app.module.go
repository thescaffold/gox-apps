package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	polylogsource "github.com/thescaffold/gox-apps-polylog/app/source"
	polylogsourcetype "github.com/thescaffold/gox-apps-polylog/app/sourcetype"
	"github.com/thescaffold/gox-apps-polylog/migrations"
	"github.com/thescaffold/gox-packages-core/module"
	polylogpipeline "github.com/thescaffold/gox-apps-polylog/pkg/pipeline"
)

var Migrations = []sql.Migration{
	&migrations.CreatePolylogSourceTypes{},
	&migrations.CreatePolylogSources{},
	&migrations.CreatePolylogSinkTypes{},
	&migrations.CreatePolylogSinks{},
	&migrations.CreatePolylogChannels{},
	&migrations.CreatePolylogEvents{},
	&migrations.CreatePolylogEventLogs{},
}

type AppModule struct{}

func (m *AppModule) Imports() []types.Module {
	return []types.Module{
		module.New(module.CoreConfig{}),
		sql.Child(&sql.Config{Migrations: Migrations}),
		&polylogsourcetype.SourceTypeModule{},
		&polylogsource.SourceModule{},
		&polylogpipeline.PipelineModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	return []any{&AppController{}, &AppService{}}
}
