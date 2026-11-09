package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	bridgelicense "github.com/thescaffold/gox-apps/libs/bridge/app/license"
	bridgelicensetype "github.com/thescaffold/gox-apps/libs/bridge/app/licensetype"
	bridgeplantype "github.com/thescaffold/gox-apps/libs/bridge/app/plantype"
	bridgepreference "github.com/thescaffold/gox-apps/libs/bridge/app/preference"
	bridgewebhook "github.com/thescaffold/gox-apps/libs/bridge/app/webhook"
	bridgewebhooklog "github.com/thescaffold/gox-apps/libs/bridge/app/webhooklog"
	"github.com/thescaffold/gox-apps/libs/bridge/migrations"
	capitalusage "github.com/thescaffold/gox-apps/libs/capital/app/usage"
	"github.com/thescaffold/gox-packages/libs/core/module"
)

var Migrations = []sql.Migration{
	&migrations.CreateBridgeLicenseTypes{},
	&migrations.CreateBridgeLicenses{},
	&migrations.CreateBridgePlanTypes{},
	&migrations.CreateBridgePreferences{},
	&migrations.CreateBridgeWebhooks{},
	&migrations.CreateBridgeWebhookLogs{},
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
		&bridgelicensetype.LicenseTypeModule{},
		&bridgelicense.LicenseModule{},
		&bridgeplantype.PlanTypeModule{},
		&bridgepreference.PreferenceModule{},
		&bridgewebhook.WebhookModule{},
		&bridgewebhooklog.WebhookLogModule{},
		&capitalusage.UsageModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	// Until the container has run, bridgeAppSvc is this svc; AppService.OnRegister
	// then replaces it with the instance the container injected (svc itself
	// may never be: the container can fill this slot with the AppService it
	// already made for AppController's field — PLAN U18). A later call must
	// not put an uninjected one back.
	svc := &AppService{}
	if bridgeAppSvc == nil {
		bridgeAppSvc = svc
	}
	return []any{&AppController{}, svc}
}
