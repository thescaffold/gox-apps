package tests

import (
	"testing"

	goosetest "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps-common/app"
	commonip "github.com/thescaffold/gox-apps-common/app/ip"
	commonproject "github.com/thescaffold/gox-apps-common/app/project"
	commonrate "github.com/thescaffold/gox-apps-common/app/rate"
	commontag "github.com/thescaffold/gox-apps-common/app/tag"
)

func TestIpEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &IpEntitySuite{}).Run()
}

type IpEntitySuite struct{ goosetest.Suite }

func (s *IpEntitySuite) TestIp_TableName() {
	s.T.Expect(commonip.Ip{}.TableName()).ToEqual("CommonIps")
}

func (s *IpEntitySuite) TestIp_Fields() {
	ip := &commonip.Ip{Value: "1.2.3.4", Country: "US"}
	s.T.Expect(ip.Value).ToEqual("1.2.3.4")
	s.T.Expect(ip.Country).ToEqual("US")
}

func TestRateEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &RateEntitySuite{}).Run()
}

type RateEntitySuite struct{ goosetest.Suite }

func (s *RateEntitySuite) TestRate_TableName() {
	s.T.Expect(commonrate.Rate{}.TableName()).ToEqual("CommonRates")
}

func (s *RateEntitySuite) TestRate_Currency() {
	r := &commonrate.Rate{Currency: "EUR"}
	s.T.Expect(r.Currency).ToEqual("EUR")
}

func TestTagEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &TagEntitySuite{}).Run()
}

type TagEntitySuite struct{ goosetest.Suite }

func (s *TagEntitySuite) TestTag_TableName() {
	s.T.Expect(commontag.Tag{}.TableName()).ToEqual("CommonTags")
}

func (s *TagEntitySuite) TestTag_Fields() {
	t := &commontag.Tag{GroupName: "apps", EntityName: "User", EntityId: "u-1", TypeId: "t-1", ServiceName: "identity"}
	s.T.Expect(t.GroupName).ToEqual("apps")
	s.T.Expect(t.EntityName).ToEqual("User")
}

func TestProjectEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &ProjectEntitySuite{}).Run()
}

type ProjectEntitySuite struct{ goosetest.Suite }

func (s *ProjectEntitySuite) TestProject_TableName() {
	s.T.Expect(commonproject.Project{}.TableName()).ToEqual("CommonProjects")
}

func TestAppModule(t *testing.T) {
	goosetest.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}

type AppModuleSuite struct{ goosetest.Suite }

func (s *AppModuleSuite) TestMigrations_Count() {
	s.T.Expect(len(app.Migrations)).ToEqual(7)
}

func (s *AppModuleSuite) TestName_IsCommon() {
	s.T.Expect(app.Name).ToEqual("common")
}

func (s *AppModuleSuite) TestSubscriptions_HasWeekly() {
	_, ok := app.Subscriptions["apps.cron.heartbeat.weekly"]
	s.T.Expect(ok).ToEqual(true)
}
