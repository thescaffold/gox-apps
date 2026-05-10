package tests

import (
	"testing"

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
	s.T.Expect(cronjob.Job{}.TableName()).ToEqual("cron_jobs")
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
	s.T.Expect(cronlog.Log{}.TableName()).ToEqual("cron_logs")
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
