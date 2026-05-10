package provider

import (
	notificationlog "github.com/thescaffold/gox-apps/libs/notification/app/log"
)

// QueuePusher dispatches a notification log for asynchronous, durable provider
// processing. Mirrors the pattern in TS app.controller.ts where new logs are
// pushed to "queue/apps/notification" with retryLimit:3, retryDelay:60.
//
// Returning an error tells the caller to fall back to synchronous Send
// (best-effort, non-durable). Pass nil to disable durable queuing — the
// default behaviour is sync Send, matching today's gox semantics.
type QueuePusher interface {
	// PushLog enqueues a notification log for processing.
	// queueName is conventionally "queue/apps/notification", jobName "message".
	PushLog(queueName, jobName string, log *notificationlog.Log, retryLimit, retryDelaySecs int) error
}

// SetQueuePusher wires (or replaces) the QueuePusher used by Send. Pass nil to
// revert to synchronous in-process dispatch. Not goroutine-safe.
func (s *ProviderService) SetQueuePusher(p QueuePusher) { s.queuePusher = p }

// dispatch is called by the notification subscription handler when a new log
// row should be processed. With a QueuePusher wired we push to the queue;
// otherwise we synchronously call Send (matches existing behaviour).
func (s *ProviderService) Dispatch(log *notificationlog.Log) {
	if s.queuePusher != nil {
		// Mirrors TS retryLimit:3, retryDelay:60.
		if err := s.queuePusher.PushLog("queue/apps/notification", "message", log, 3, 60); err == nil {
			return
		}
	}
	s.Send(log)
}
