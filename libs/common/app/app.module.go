package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	commonip "github.com/thescaffold/gox-apps/libs/common/app/ip"
	commonproject "github.com/thescaffold/gox-apps/libs/common/app/project"
	commonprojecttype "github.com/thescaffold/gox-apps/libs/common/app/projecttype"
	commonrate "github.com/thescaffold/gox-apps/libs/common/app/rate"
	commonratelog "github.com/thescaffold/gox-apps/libs/common/app/ratelog"
	commontag "github.com/thescaffold/gox-apps/libs/common/app/tag"
	commontagtype "github.com/thescaffold/gox-apps/libs/common/app/tagtype"
	"github.com/thescaffold/gox-apps/libs/common/migrations"
	"github.com/thescaffold/gox-packages/libs/core/module"
)

var Migrations = []sql.Migration{
	&migrations.CreateCommonIps{},
	&migrations.CreateCommonRates{},
	&migrations.CreateCommonRateLogs{},
	&migrations.CreateCommonTagTypes{},
	&migrations.CreateCommonTags{},
	&migrations.CreateCommonProjectTypes{},
	&migrations.CreateCommonProjects{},
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
		&commonip.IpModule{},
		&commonrate.RateModule{},
		&commonratelog.RateLogModule{},
		&commontag.TagModule{},
		&commontagtype.TagTypeModule{},
		&commonproject.ProjectModule{},
		&commonprojecttype.ProjectTypeModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	svc := &AppService{}
	commonAppSvc = svc
	return []any{&AppController{}, svc}
}
