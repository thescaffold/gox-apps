package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	fusshist "github.com/thescaffold/gox-apps/libs/fuss/app/history"
	fusstok "github.com/thescaffold/gox-apps/libs/fuss/app/token"
	fusstoklog "github.com/thescaffold/gox-apps/libs/fuss/app/tokenlog"
	"github.com/thescaffold/gox-apps/libs/fuss/migrations"
	"github.com/thescaffold/gox-packages/libs/core/module"
)

var Migrations = []sql.Migration{
	&migrations.CreateFussTokens{},
	&migrations.CreateFussTokenLogs{},
	&migrations.CreateFussHistories{},
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
		&fusstok.TokenModule{},
		&fusstoklog.TokenLogModule{},
		&fusshist.HistoryModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	// Until the container has run, fussAppSvc is this svc; AppService.OnRegister
	// then replaces it with the instance the container injected (svc itself
	// may never be: the container can fill this slot with the AppService it
	// already made for AppController's field — PLAN U18). A later call must
	// not put an uninjected one back.
	svc := &AppService{}
	if fussAppSvc == nil {
		fussAppSvc = svc
	}
	return []any{&AppController{}, svc}
}
