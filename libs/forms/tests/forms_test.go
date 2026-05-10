package tests

import (
	"testing"

	goosetest "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps/libs/forms/app"
	formsform "github.com/thescaffold/gox-apps/libs/forms/app/form"
	formsformfield "github.com/thescaffold/gox-apps/libs/forms/app/formfield"
	formsformlog "github.com/thescaffold/gox-apps/libs/forms/app/formlog"
	formsformtype "github.com/thescaffold/gox-apps/libs/forms/app/formtype"
)

func TestFormTypeEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &FormTypeEntitySuite{}).Run()
}

type FormTypeEntitySuite struct{ goosetest.Suite }

func (s *FormTypeEntitySuite) TestTableName() {
	s.T.Expect(formsformtype.FormType{}.TableName()).ToEqual("FormFormTypes")
}

func TestFormEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &FormEntitySuite{}).Run()
}

type FormEntitySuite struct{ goosetest.Suite }

func (s *FormEntitySuite) TestTableName() {
	s.T.Expect(formsform.Form{}.TableName()).ToEqual("FormForms")
}

func TestFormFieldEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &FormFieldEntitySuite{}).Run()
}

type FormFieldEntitySuite struct{ goosetest.Suite }

func (s *FormFieldEntitySuite) TestTableName() {
	s.T.Expect(formsformfield.FormField{}.TableName()).ToEqual("FormFormFields")
}

func TestFormLogEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &FormLogEntitySuite{}).Run()
}

type FormLogEntitySuite struct{ goosetest.Suite }

func (s *FormLogEntitySuite) TestTableName() {
	s.T.Expect(formsformlog.FormLog{}.TableName()).ToEqual("FormFormLogs")
}

func TestAppModule(t *testing.T) {
	goosetest.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}

type AppModuleSuite struct{ goosetest.Suite }

func (s *AppModuleSuite) TestMigrations_Count() {
	s.T.Expect(len(app.Migrations)).ToEqual(4)
}

func (s *AppModuleSuite) TestName() {
	s.T.Expect(app.Name).ToEqual("forms")
}

func (s *AppModuleSuite) TestSubscriptions_Empty() {
	// Mirrors TS subscriptions = [].
	s.T.Expect(len(app.Subscriptions)).ToEqual(0)
}
