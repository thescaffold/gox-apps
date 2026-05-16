package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	auditlog "github.com/thescaffold/gox-apps/libs/audit/app/log"
	"github.com/thescaffold/gox-apps/libs/audit/migrations"
	"github.com/thescaffold/gox-packages/libs/core/module"
)

var Migrations = []sql.Migration{
	&migrations.CreateAuditLogs{},
}

type AppModule struct{}

func (m *AppModule) Imports() []types.Module {
	return []types.Module{
		// TranslationPaths mirror TS index.ts `paths = [translations/en/ntx/apps/${name}.yaml]`;
		// CoreModule.Boot pre-loads them so translate() resolves real strings.
		module.New(module.CoreConfig{
			TranslationPaths: Paths,
		}),
		sql.Child(&sql.Config{
			Migrations: Migrations,
		}),
		&auditlog.LogModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any {
	return []any{&AppService{}}
}

// AppController is declared so goose registers it and the routes referencing
// AppController{} resolve to the same DI-managed instance.
func (m *AppModule) Declarations() []any {
	svc := &AppService{}
	appSvc = svc
	return []any{&AppController{}, svc}
}
