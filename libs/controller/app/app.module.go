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
		// TranslationPaths mirror TS index.ts `paths = [translations/en/ntx/apps/${name}.yaml]`;
		// CoreModule.Boot pre-loads them into MediaService before the first request,
		// matching the TS app bootstrap calling mediaService.load(paths).
		module.New(module.CoreConfig{
			TranslationPaths: Paths,
		}),
		sql.Child(&sql.Config{Migrations: Migrations}),
		&ctrlroute.RouteModule{},
		&ctrlreq.RequestModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	// Until the container has run, controllerAppSvc is this svc; AppService.OnRegister
	// then replaces it with the instance the container injected (svc itself
	// may never be: the container can fill this slot with the AppService it
	// already made for AppController's field — PLAN U18). A later call must
	// not put an uninjected one back.
	svc := &AppService{}
	if controllerAppSvc == nil {
		controllerAppSvc = svc
	}
	return []any{&AppController{}, svc}
}
