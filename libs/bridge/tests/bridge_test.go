package tests

import (
	"testing"

	goosetest "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps/libs/bridge/app"
	bridgelicense "github.com/thescaffold/gox-apps/libs/bridge/app/license"
	bridgelicensetype "github.com/thescaffold/gox-apps/libs/bridge/app/licensetype"
)

func TestLicenseTypeEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &LicenseTypeEntitySuite{}).Run()
}

type LicenseTypeEntitySuite struct{ goosetest.Suite }

func (s *LicenseTypeEntitySuite) TestTableName() {
	s.T.Expect(bridgelicensetype.LicenseType{}.TableName()).ToEqual("BridgeLicenseTypes")
}

func TestLicenseEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &LicenseEntitySuite{}).Run()
}

type LicenseEntitySuite struct{ goosetest.Suite }

func (s *LicenseEntitySuite) TestTableName() {
	s.T.Expect(bridgelicense.License{}.TableName()).ToEqual("BridgeLicenses")
}

func TestAppModule(t *testing.T) {
	goosetest.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}

type AppModuleSuite struct{ goosetest.Suite }

func (s *AppModuleSuite) TestMigrations_Count() {
	s.T.Expect(len(app.Migrations)).ToEqual(6)
}

func (s *AppModuleSuite) TestName() {
	s.T.Expect(app.Name).ToEqual("bridge")
}
