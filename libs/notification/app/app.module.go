package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	notificationlog "github.com/thescaffold/gox-apps/libs/notification/app/log"
	notificationmessage "github.com/thescaffold/gox-apps/libs/notification/app/message"
	notificationrule "github.com/thescaffold/gox-apps/libs/notification/app/rule"
	notificationtemplate "github.com/thescaffold/gox-apps/libs/notification/app/template"
	"github.com/thescaffold/gox-apps/libs/notification/migrations"
	notificationprovider "github.com/thescaffold/gox-apps/libs/notification/pkg/provider"
	"github.com/thescaffold/gox-packages/libs/core/module"
)

var Migrations = []sql.Migration{
	&migrations.CreateNotificationTemplates{},
	&migrations.CreateNotificationRules{},
	&migrations.CreateNotificationMessages{},
	&migrations.CreateNotificationLogs{},
	&migrations.AddLinkToNotificationMessages{},
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
		&notificationlog.LogModule{},
		&notificationrule.RuleModule{},
		&notificationtemplate.TemplateModule{},
		&notificationmessage.MessageModule{},
		&notificationprovider.ProviderModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	// Until the container has run, notificationAppSvc is this svc; AppService.OnRegister
	// then replaces it with the instance the container injected (svc itself
	// may never be: the container can fill this slot with the AppService it
	// already made for AppController's field — PLAN U18). A later call must
	// not put an uninjected one back.
	svc := &AppService{}
	if notificationAppSvc == nil {
		notificationAppSvc = svc
	}
	return []any{&AppController{}, svc}
}
