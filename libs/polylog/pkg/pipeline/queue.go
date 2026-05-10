package pipeline

import (
	polyloevent "github.com/thescaffold/gox-apps-polylog/app/event"
)

// QueuePusher dispatches an event for asynchronous, durable processing.
// Implementations typically POST to /apps/queue/push so unflushed events
// survive a process crash. Mirrors TS queueAppService.push() semantics:
// {queue, job, data, retryLimit, retryDelay}.
//
// Returning an error tells the caller to fall back to in-process goroutine
// processing (best-effort, non-durable). The default Pipeline behaviour when
// QueuePusher is nil is the in-process goroutine — no behaviour change.
type QueuePusher interface {
	// PushEvent enqueues an event for processing by the polylog pipeline.
	// queueName is conventionally "queue/apps/polylog", jobName "ingest".
	PushEvent(queueName, jobName string, event *polyloevent.Event, retryLimit, retryDelaySecs int) error
}

// SetQueuePusher wires (or replaces) the QueuePusher used by Ingest. Pass nil
// to revert to in-process goroutine dispatch. Not goroutine-safe.
func (s *PipelineService) SetQueuePusher(p QueuePusher) { s.queuePusher = p }

// dispatch is called from Ingest when the event should be processed
// asynchronously. When a QueuePusher is wired and accepts the push, the
// event is durably queued. Otherwise we fall back to the legacy goroutine.
func (s *PipelineService) dispatch(ev *polyloevent.Event) {
	if s.queuePusher != nil {
		// Mirrors TS retryLimit:3, retryDelay:60 (1 minute).
		if err := s.queuePusher.PushEvent("queue/apps/polylog", "ingest", ev, 3, 60); err == nil {
			return
		}
		// Pusher rejected the event; fall through to in-process processing
		// rather than dropping it.
	}
	go func() { _ = s.Process(ev) }()
}
