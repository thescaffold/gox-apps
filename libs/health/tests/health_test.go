package tests

import (
	"testing"

	test "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps-health/app"
	healthlog "github.com/thescaffold/gox-apps-health/app/log"
	healthsvc "github.com/thescaffold/gox-apps-health/app/service"
	healthsum "github.com/thescaffold/gox-apps-health/app/summary"
)

func TestLogEntity(t *testing.T) {
	test.NewSuiteRunner(t, &LogEntitySuite{}).Run()
}

type LogEntitySuite struct{ test.Suite }

func (s *LogEntitySuite) TestLog_Initialization() {
	l := &healthlog.Log{ServiceId: "svc-1"}
	s.T.Expect(l.ServiceId).ToEqual("svc-1")
}

func (s *LogEntitySuite) TestLog_TableName() {
	s.T.Expect(healthlog.Log{}.TableName()).ToEqual("HealthLogs")
}

func TestServiceEntity(t *testing.T) {
	test.NewSuiteRunner(t, &ServiceEntitySuite{}).Run()
}

type ServiceEntitySuite struct{ test.Suite }

func (s *ServiceEntitySuite) TestService_Initialization() {
	svc := &healthsvc.Service{Name: "api"}
	s.T.Expect(svc.Name).ToEqual("api")
}

func (s *ServiceEntitySuite) TestService_TableName() {
	s.T.Expect(healthsvc.Service{}.TableName()).ToEqual("HealthServices")
}

func TestSummaryEntity(t *testing.T) {
	test.NewSuiteRunner(t, &SummaryEntitySuite{}).Run()
}

type SummaryEntitySuite struct{ test.Suite }

func (s *SummaryEntitySuite) TestSummary_Initialization() {
	sum := &healthsum.Summary{ServiceId: "svc-1", Type: "uptime"}
	s.T.Expect(sum.ServiceId).ToEqual("svc-1")
	s.T.Expect(sum.Type).ToEqual("uptime")
}

func (s *SummaryEntitySuite) TestSummary_TableName() {
	s.T.Expect(healthsum.Summary{}.TableName()).ToEqual("HealthSummaries")
}

func TestAppModule(t *testing.T) {
	test.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}

type AppModuleSuite struct{ test.Suite }

func (s *AppModuleSuite) TestMigrations_NotEmpty() {
	s.T.Expect(len(app.Migrations)).ToEqual(3)
}

func (s *AppModuleSuite) TestName_IsHealth() {
	s.T.Expect(app.Name).ToEqual("health")
}

func (s *AppModuleSuite) TestSubscriptions_Count() {
	s.T.Expect(len(app.Subscriptions)).ToEqual(3)
}
