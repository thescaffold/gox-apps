package app

import (
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	formsform "github.com/thescaffold/gox-apps/libs/forms/app/form"
	formsformfield "github.com/thescaffold/gox-apps/libs/forms/app/formfield"
	formsformlog "github.com/thescaffold/gox-apps/libs/forms/app/formlog"
	formsformtype "github.com/thescaffold/gox-apps/libs/forms/app/formtype"
	"github.com/thescaffold/gox-apps/libs/forms/migrations"
	"github.com/thescaffold/gox-packages/libs/core/module"
)

var Migrations = []sql.Migration{
	&migrations.CreateFormFormTypes{},
	&migrations.CreateFormForms{},
	&migrations.CreateFormFormFields{},
	&migrations.CreateFormFormLogs{},
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
		&formsformtype.FormTypeModule{},
		&formsform.FormModule{},
		&formsformfield.FormFieldModule{},
		&formsformlog.FormLogModule{},
		ROUTES,
	}
}

func (m *AppModule) Exports() []any { return []any{&AppService{}} }

func (m *AppModule) Declarations() []any {
	return []any{&AppController{}, &AppService{}}
}
