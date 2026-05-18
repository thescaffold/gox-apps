package app

import (
	"context"
	"time"

	goqueues "github.com/awesome-goose/goose/modules/queues"
	queuejob "github.com/thescaffold/gox-apps/libs/queue/app/job"
	queuelog "github.com/thescaffold/gox-apps/libs/queue/app/log"
)

// AppService is the producer + consumer surface for the queue lib. The
// producer half (Push/Pop/Log) delegates to the underlying goose `Queue`.
// The consumer half (RegisterJob/RegisterSimpleJob) routes handler
// registration through goose's auto-scaling worker pool — `goqueues.Queue.
// Process` is idempotent (it calls `Initialize` internally) so handlers can
// be registered from each app module's Boot hook without coordination.
type AppService struct {
	queueSvc *goqueues.Queue    `inject:""`
	jobs     *queuejob.JobEntity `inject:""`
	logs     *queuelog.LogEntity `inject:""`
}

// Housekeep deletes successful QueueJobs and QueueLogs older than cutoff.
// Mirrors TS AppController.subscriptions['apps.cron.heartbeat.hourly'] which
// deletes Job and Log rows where status=Success AND createdAt < threshold.
func (s *AppService) Housekeep(cutoff time.Time) {
	if s.jobs != nil {
		_, _ = s.jobs.Delete(`"status" = ? AND "created_at" < ?`, "success", cutoff)
	}
	if s.logs != nil {
		_, _ = s.logs.Delete(`"status" = ? AND "created_at" < ?`, "success", cutoff)
	}
}

func (s *AppService) GetHello() string { return "Hello World!" }

// Push enqueues a new job. Mirrors TS `queueAppService.push(queue, job, data, opts)`.
func (s *AppService) Push(queueName, jobName string, data any, config *goqueues.JobConfig) (*goqueues.QueueJob, error) {
	return s.queueSvc.Push(queueName, jobName, data, config)
}

// Pop fetches the next pending job for the given queue/job pair. Producers
// typically don't call this — workers do, via the handlers registered by
// RegisterJob — but it's exposed for tests and manual drain.
func (s *AppService) Pop(queueName, jobName string) (*goqueues.QueueJob, error) {
	return s.queueSvc.Pop(queueName, jobName)
}

// Log records a job's final status (success/failed/timeout) and stores the
// handler's return value as the job's `output` column.
func (s *AppService) Log(jobId, status string, output any) (*goqueues.QueueJob, error) {
	return s.queueSvc.Log(jobId, status, output)
}

// RegisterJob registers a handler with goose's worker pool. The handler is
// processed by an auto-scaling set of goroutines (configurable via
// `goqueues.JobHandler.WithMinWorkers/WithMaxWorkers/WithPollInterval/
// WithTimeout`). Safe to call multiple times — `goqueues.Queue.Process`
// already calls `Initialize` internally and supports concurrent registrations.
//
// Typical use: each app's `app.module.go` injects `*queue.AppService` and
// calls RegisterJob from its OnRegister/Boot hook, fed by the app's
// `Jobs []*goqueues.JobHandler` slice declared in `app/index.go`.
func (s *AppService) RegisterJob(handler *goqueues.JobHandler) {
	if handler == nil || s.queueSvc == nil {
		return
	}
	s.queueSvc.Process(handler)
}

// RegisterJobs is the bulk variant of RegisterJob — registers every handler
// in the slice. Skips nil entries silently. Useful for wiring the
// `app.Jobs` slice from each gox-app's index.go in a single call.
func (s *AppService) RegisterJobs(handlers []*goqueues.JobHandler) {
	for _, h := range handlers {
		s.RegisterJob(h)
	}
}

// RegisterSimpleJob is the ergonomic wrapper apps reach for first: provide a
// queue name, job name, and a `func(*QueueJob) (any, error)` and goose
// builds the JobHandler with sensible defaults (MinWorkers=1, MaxWorkers=4,
// PollInterval=3s, IdleThreshold=3, Timeout=30s, no backoff). Apps that need
// the full knob set should build a `*goqueues.JobHandler` themselves and call
// RegisterJob.
func (s *AppService) RegisterSimpleJob(queueName, jobName string, fn func(job *goqueues.QueueJob) (any, error)) {
	s.RegisterJob(goqueues.NewSimpleHandler(queueName, jobName, fn))
}

// RegisterContextJob is the context-aware sibling of RegisterSimpleJob, for
// handlers that need cancellation/timeout propagation. The supplied `fn`
// receives goose's per-job context.
func (s *AppService) RegisterContextJob(queueName, jobName string, fn func(ctx context.Context, job *goqueues.QueueJob) (any, error)) {
	s.RegisterJob(goqueues.NewHandler(queueName, jobName, fn))
}
