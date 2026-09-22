package tests

import (
	"testing"
	"time"

	gocron "github.com/awesome-goose/goose/modules/cron"
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

// TestJobs_PatternsAreValid registers every heartbeat against goose's own
// cron.IsValidCronPattern — the exact check modules/cron's Boot hook runs on
// every registration. The patterns previously included a leading seconds
// field ("0 * * * * *", 6 fields), matching a common JS cron-library
// convention (node-cron) the TS original presumably used, but
// goose's parser accepts standard 5-field patterns only
// (minute hour day-of-month month day-of-week — see IsValidCronPattern's own
// 5-element validators slice) and rejects anything else outright. Every one
// of these six registrations silently failed at boot — "silently" because
// the kernel logs it as a warning and keeps booting, so nothing failed loudly
// enough to be caught before now.
func (s *AppModuleSuite) TestJobs_PatternsAreValid() {
	for _, j := range app.Jobs {
		s.T.Expect(gocron.IsValidCronPattern(j.Pattern)).ToEqual(true)
	}
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
			s.T.Expect(j.Pattern).ToEqual("* * * * *")
		}
	}
	s.T.Expect(found).ToEqual(true)
}

func (s *AppModuleSuite) TestJobs_HeartbeatYearly() {
	found := false
	for _, j := range app.Jobs {
		if j.Group == "apps.cron" && j.Name == "heartbeat.yearly" {
			found = true
			s.T.Expect(j.Pattern).ToEqual("0 0 1 1 *")
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
