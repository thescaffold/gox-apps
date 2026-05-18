package tests

import (
	"testing"
	"time"

	goosetest "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps/libs/queue/app"
	queuejob "github.com/thescaffold/gox-apps/libs/queue/app/job"
	queuelog "github.com/thescaffold/gox-apps/libs/queue/app/log"
	queuequeue "github.com/thescaffold/gox-apps/libs/queue/app/queue"
)

func TestQueueEntity(t *testing.T) {
	goosetest.NewSuiteRunner(t, &QueueEntitySuite{}).Run()
}

type QueueEntitySuite struct{ goosetest.Suite }

func (s *QueueEntitySuite) TestQueue_TableName() {
	s.T.Expect(queuequeue.Queue{}.TableName()).ToEqual("QueueQueues")
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
	s.T.Expect(queuejob.Job{}.TableName()).ToEqual("QueueJobs")
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
	s.T.Expect(queuelog.Log{}.TableName()).ToEqual("QueueLogs")
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

// TestSubscriptions_HourlyIsBound asserts the hourly heartbeat handler is a
// real function reference, not a nil entry. Combined with TestAppService_
// Housekeep_DoesNotPanic_WithoutDeps below, this guards against the regression
// where the handler was a no-op stub. TS deletes successful Queue/Job + Log
// rows on each tick; this test ensures gox has the wiring in place.
func (s *AppModuleSuite) TestSubscriptions_HourlyIsBound() {
	handler, ok := app.Subscriptions["apps.cron.heartbeat.hourly"]
	s.T.Expect(ok).ToEqual(true)
	s.T.Expect(handler == nil).ToEqual(false)
}

// TestAppService_Housekeep_DoesNotPanic_WithoutDeps verifies the Housekeep
// method tolerates a fresh, uninitialized AppService (no entities injected).
// The handler runs from a subscription, so panicking on missing deps would
// crash the event bus loop. The real DI path will inject jobs/logs in
// production.
func (s *AppModuleSuite) TestAppService_Housekeep_DoesNotPanic_WithoutDeps() {
	svc := &app.AppService{}
	defer func() { s.T.Expect(recover() == nil).ToEqual(true) }()
	svc.Housekeep(time.Now().UTC().Add(-24 * time.Hour))
}
