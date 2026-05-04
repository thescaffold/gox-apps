package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	bridgelicense "github.com/thescaffold/gox-apps-bridge/app/license"
	bridgelicensetype "github.com/thescaffold/gox-apps-bridge/app/licensetype"
	bridgeplantype "github.com/thescaffold/gox-apps-bridge/app/plantype"
	bridgepreference "github.com/thescaffold/gox-apps-bridge/app/preference"
	bridgewebhook "github.com/thescaffold/gox-apps-bridge/app/webhook"
	bridgewebhooklog "github.com/thescaffold/gox-apps-bridge/app/webhooklog"
	"github.com/thescaffold/gox-apps-bridge/migrations"
	"github.com/thescaffold/gox-packages-core/module"
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
		module.New(module.CoreConfig{}),
		sql.Child(&sql.Config{Migrations: Migrations}),
		&bridgelicensetype.LicenseTypeModule{},
		&bridgelicense.LicenseModule{},
		&bridgeplantype.PlanTypeModule{},
		&bridgepreference.PreferenceModule{},
		&bridgewebhook.WebhookModule{},
		&bridgewebhooklog.WebhookLogModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	return []any{&AppController{}, &AppService{}}
}
