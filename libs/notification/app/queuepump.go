package app

import (
	goqueues "github.com/awesome-goose/goose/modules/queues"
	notificationlog "github.com/thescaffold/gox-apps/libs/notification/app/log"
	notificationprovider "github.com/thescaffold/gox-apps/libs/notification/pkg/provider"
	queueapp "github.com/thescaffold/gox-apps/libs/queue/app"
)

// queuePump adapts queueapp.AppService to the provider.QueuePusher contract.
// Used by AppController.OnRegister to wire durable dispatch into the
// ProviderService — `dispatch` pushes to `queue/apps/notification/message`
// rather than running Send synchronously.
type queuePump struct {
	queue *queueapp.AppService
}

// Compile-time interface check.
var _ notificationprovider.QueuePusher = (*queuePump)(nil)

// PushLog enqueues a notification Log row for the worker. The worker
// (notification.Jobs[0] from index.go) drains the queue and calls
// ProviderService.Send via AppService.OnMessageJob.
func (p *queuePump) PushLog(queueName, jobName string, log *notificationlog.Log, retryLimit, retryDelaySecs int) error {
	if p.queue == nil || log == nil {
		return errNoQueue
	}
	cfg := &goqueues.JobConfig{
		RetryLimit: retryLimit,
		RetryDelay: retryDelaySecs * 1000, // seconds → ms (goose convention)
	}
	_, err := p.queue.Push(queueName, jobName, map[string]any{
		"log":       log,
		"key":       log.Key,
		"reference": log.Reference,
	}, cfg)
	return err
}

// errNoQueue signals to Dispatch that no queue is wired so it should fall
// back to synchronous Send.
var errNoQueue = errString("notification: queueapp not configured")

type errString string

func (e errString) Error() string { return string(e) }
