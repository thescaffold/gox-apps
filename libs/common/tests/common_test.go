package tests

import (
	"testing"

	goosetest "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps/libs/common/app"
	commonip "github.com/thescaffold/gox-apps/libs/common/app/ip"
	commonproject "github.com/thescaffold/gox-apps/libs/common/app/project"
	commonrate "github.com/thescaffold/gox-apps/libs/common/app/rate"
	commontag "github.com/thescaffold/gox-apps/libs/common/app/tag"
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

func TestLocationIP(t *testing.T) {
	goosetest.NewSuiteRunner(t, &LocationIPSuite{}).Run()
}

type LocationIPSuite struct{ goosetest.Suite }

func (s *LocationIPSuite) TestQueryWins() {
	dto := &app.GetLocationDto{Ip: "203.0.113.9", HeaderIP: "198.51.100.1"}
	s.T.Expect(app.LocationIP(dto)).ToEqual("203.0.113.9")
}

func (s *LocationIPSuite) TestFirstForwardedHop() {
	dto := &app.GetLocationDto{HeaderIP: "198.51.100.1, 10.0.0.2"}
	s.T.Expect(app.LocationIP(dto)).ToEqual("198.51.100.1")
}

func (s *LocationIPSuite) TestContextAddress() {
	dto := &app.GetLocationDto{}
	dto.Ctx.IP = "192.0.2.7"
	s.T.Expect(app.LocationIP(dto)).ToEqual("192.0.2.7")
}

func (s *LocationIPSuite) TestLoopbackWhenNothingKnown() {
	s.T.Expect(app.LocationIP(&app.GetLocationDto{})).ToEqual("127.0.0.1")
}
