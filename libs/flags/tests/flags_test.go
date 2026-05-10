package tests

import (
	"testing"

	goosetest "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps-flags/app"
	flagsenvironment "github.com/thescaffold/gox-apps-flags/app/environment"
	flagsenvironmenttype "github.com/thescaffold/gox-apps-flags/app/environmenttype"
	flagsflag "github.com/thescaffold/gox-apps-flags/app/flag"
	flagsflaglog "github.com/thescaffold/gox-apps-flags/app/flaglog"
)

func TestEnvironmentTypeEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &EnvironmentTypeEntitySuite{}).Run()
}

type EnvironmentTypeEntitySuite struct{ goosetest.Suite }

func (s *EnvironmentTypeEntitySuite) TestTableName() {
	s.T.Expect(flagsenvironmenttype.EnvironmentType{}.TableName()).ToEqual("FlagEnvironmentTypes")
}

func TestEnvironmentEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &EnvironmentEntitySuite{}).Run()
}

type EnvironmentEntitySuite struct{ goosetest.Suite }

func (s *EnvironmentEntitySuite) TestTableName() {
	s.T.Expect(flagsenvironment.Environment{}.TableName()).ToEqual("FlagEnvironments")
}

func TestFlagEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &FlagEntitySuite{}).Run()
}

type FlagEntitySuite struct{ goosetest.Suite }

func (s *FlagEntitySuite) TestTableName() {
	s.T.Expect(flagsflag.Flag{}.TableName()).ToEqual("FlagFlags")
}

func TestFlagLogEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &FlagLogEntitySuite{}).Run()
}

type FlagLogEntitySuite struct{ goosetest.Suite }

func (s *FlagLogEntitySuite) TestTableName() {
	s.T.Expect(flagsflaglog.FlagLog{}.TableName()).ToEqual("FlagFlagLogs")
}

func TestAppModule(t *testing.T) {
	goosetest.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}

type AppModuleSuite struct{ goosetest.Suite }

func (s *AppModuleSuite) TestMigrations_Count() {
	s.T.Expect(len(app.Migrations)).ToEqual(4)
}

func (s *AppModuleSuite) TestName() {
	s.T.Expect(app.Name).ToEqual("flags")
}

func (s *AppModuleSuite) TestSubscriptions_HasFlagEvents() {
	// Mirrors TS subscriptions: 'apps.flags.flag.register', 'apps.flags.flag.log'.
	_, hasReg := app.Subscriptions["apps.flags.flag.register"]
	_, hasLog := app.Subscriptions["apps.flags.flag.log"]
	s.T.Expect(hasReg).ToEqual(true)
	s.T.Expect(hasLog).ToEqual(true)
}
