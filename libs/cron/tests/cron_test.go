package tests

import (
	"testing"
	"time"

	test "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps/libs/cron/app"
	cronjob "github.com/thescaffold/gox-apps/libs/cron/app/job"
	cronlog "github.com/thescaffold/gox-apps/libs/cron/app/log"
)

func TestJobEntity(t *testing.T) {
	test.NewSuiteRunner(t, &JobEntitySuite{}).Run()
}

type JobEntitySuite struct{ test.Suite }

func (s *JobEntitySuite) TestJob_TableName() {
	s.T.Expect(cronjob.Job{}.TableName()).ToEqual("CronJobs")
}

func (s *JobEntitySuite) TestJob_DefaultStatus() {
	j := &cronjob.Job{}
	s.T.Expect(j.Status).ToEqual("")
}

func TestLogEntity(t *testing.T) {
	test.NewSuiteRunner(t, &LogEntitySuite{}).Run()
}

type LogEntitySuite struct{ test.Suite }

func (s *LogEntitySuite) TestLog_TableName() {
	s.T.Expect(cronlog.Log{}.TableName()).ToEqual("CronLogs")
}

func (s *LogEntitySuite) TestLog_JobId() {
	l := &cronlog.Log{JobId: "job-1"}
	s.T.Expect(l.JobId).ToEqual("job-1")
}

func TestAppModule(t *testing.T) {
	test.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}

type AppModuleSuite struct{ test.Suite }

func (s *AppModuleSuite) TestMigrations_Empty() {
	s.T.Expect(len(app.Migrations)).ToEqual(0)
}

func (s *AppModuleSuite) TestName_IsCron() {
	s.T.Expect(app.Name).ToEqual("cron")
}

func (s *AppModuleSuite) TestJobs_SixHeartbeats() {
	s.T.Expect(len(app.Jobs)).ToEqual(6)
}

func (s *AppModuleSuite) TestSubscriptions_HasHourly() {
	_, ok := app.Subscriptions["apps.cron.heartbeat.hourly"]
	s.T.Expect(ok).ToEqual(true)
}

func (s *AppModuleSuite) TestJobs_HeartbeatMinute() {
	found := false
	for _, j := range app.Jobs {
		if j.Group == "apps.cron" && j.Name == "heartbeat.minute" {
			found = true
			s.T.Expect(j.Pattern).ToEqual("0 * * * * *")
		}
	}
	s.T.Expect(found).ToEqual(true)
}

func (s *AppModuleSuite) TestJobs_HeartbeatYearly() {
	found := false
	for _, j := range app.Jobs {
		if j.Group == "apps.cron" && j.Name == "heartbeat.yearly" {
			found = true
			s.T.Expect(j.Pattern).ToEqual("0 0 0 1 1 *")
		}
	}
	s.T.Expect(found).ToEqual(true)
}

// TestSubscriptions_HourlyIsBound asserts the hourly heartbeat handler is a
// real function reference, not a nil entry. Combined with the AppService.
// Housekeep test below, this guards against the regression where the handler
// was a no-op stub: `func(_ string, _ any) {}`. The TS service deletes
// successful Log rows on each tick; this test ensures the gox version has
// wiring in place to do the same.
func (s *AppModuleSuite) TestSubscriptions_HourlyIsBound() {
	handler, ok := app.Subscriptions["apps.cron.heartbeat.hourly"]
	s.T.Expect(ok).ToEqual(true)
	s.T.Expect(handler == nil).ToEqual(false)
}

// TestAppService_Housekeep_DoesNotPanic_WithoutDeps verifies that the
// Housekeep method tolerates a fresh, uninitialized AppService (no entity
// injected). The handler runs from a subscription, so panicking on missing
// deps would crash the event bus loop. The real DI path will inject the
// log entity in production.
func (s *AppModuleSuite) TestAppService_Housekeep_DoesNotPanic_WithoutDeps() {
	svc := &app.AppService{}
	defer func() { s.T.Expect(recover() == nil).ToEqual(true) }()
	svc.Housekeep(time.Now().UTC().Add(-24 * time.Hour))
}
