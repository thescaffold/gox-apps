package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	polylogchannel "github.com/thescaffold/gox-apps/libs/polylog/app/channel"
	polylogconfig "github.com/thescaffold/gox-apps/libs/polylog/app/config"
	polylogevent "github.com/thescaffold/gox-apps/libs/polylog/app/event"
	polylogeventlog "github.com/thescaffold/gox-apps/libs/polylog/app/eventlog"
	polylogsink "github.com/thescaffold/gox-apps/libs/polylog/app/sink"
	polylogsinktype "github.com/thescaffold/gox-apps/libs/polylog/app/sinktype"
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
	// Phase 4 schema alignment — adds `entity` column on PolylogEvents so
	// the TS Event.entityName column-name mapping resolves correctly.
	&migrations.AlignPolylogEventEntity{},
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
		&polylogsourcetype.SourceTypeModule{},
		&polylogsource.SourceModule{},
		&polylogsinktype.SinkTypeModule{},
		&polylogsink.SinkModule{},
		&polylogchannel.ChannelModule{},
		&polylogevent.EventModule{},
		&polylogeventlog.EventLogModule{},
		&polylogconfig.ConfigModule{},
		&polylogpipeline.PipelineModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	// Until the container has run, polylogAppSvc is this svc; AppService.OnRegister
	// then replaces it with the instance the container injected (svc itself
	// may never be: the container can fill this slot with the AppService it
	// already made for AppController's field — PLAN U18). A later call must
	// not put an uninjected one back.
	svc := &AppService{}
	if polylogAppSvc == nil {
		polylogAppSvc = svc
	}
	return []any{&AppController{}, svc}
}
