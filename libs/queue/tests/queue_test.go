package tests

import (
	"testing"

	goosetest "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps-queue/app"
	queuejob "github.com/thescaffold/gox-apps-queue/app/job"
	queuelog "github.com/thescaffold/gox-apps-queue/app/log"
	queuequeue "github.com/thescaffold/gox-apps-queue/app/queue"
)

func TestQueueEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &QueueEntitySuite{}).Run()
}

type QueueEntitySuite struct{ goosetest.Suite }

func (s *QueueEntitySuite) TestQueue_TableName() {
	s.T.Expect(queuequeue.Queue{}.TableName()).ToEqual("queue_queues")
}

func (s *QueueEntitySuite) TestQueue_DefaultStatus() {
	q := &queuequeue.Queue{}
	s.T.Expect(q.Status).ToEqual("")
}

func TestJobEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &JobEntitySuite{}).Run()
}

type JobEntitySuite struct{ goosetest.Suite }

func (s *JobEntitySuite) TestJob_TableName() {
	s.T.Expect(queuejob.Job{}.TableName()).ToEqual("queue_jobs")
}

func (s *JobEntitySuite) TestJob_QueueId() {
	j := &queuejob.Job{QueueId: "q-1"}
	s.T.Expect(j.QueueId).ToEqual("q-1")
}

func TestLogEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &LogEntitySuite{}).Run()
}

type LogEntitySuite struct{ goosetest.Suite }

func (s *LogEntitySuite) TestLog_TableName() {
	s.T.Expect(queuelog.Log{}.TableName()).ToEqual("queue_logs")
}

func (s *LogEntitySuite) TestLog_JobId() {
	l := &queuelog.Log{JobId: "j-1"}
	s.T.Expect(l.JobId).ToEqual("j-1")
}

func TestAppModule(t *testing.T) {
	goosetest.NewSuiteRunner(t, &AppModuleSuite{}).Run()
}

type AppModuleSuite struct{ goosetest.Suite }

func (s *AppModuleSuite) TestMigrations_Empty() {
	s.T.Expect(len(app.Migrations)).ToEqual(0)
}

func (s *AppModuleSuite) TestName_IsQueue() {
	s.T.Expect(app.Name).ToEqual("queue")
}

func (s *AppModuleSuite) TestSubscriptions_HasHourly() {
	_, ok := app.Subscriptions["apps.cron.heartbeat.hourly"]
	s.T.Expect(ok).ToEqual(true)
}
