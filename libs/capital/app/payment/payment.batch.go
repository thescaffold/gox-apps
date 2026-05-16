package payment

import (
	"time"

	goqueues "github.com/awesome-goose/goose/modules/queues"
	capitalplan "github.com/thescaffold/gox-apps/libs/capital/app/plan"
	queueapp "github.com/thescaffold/gox-apps/libs/queue/app"
)

// BatchPusher is the queue-producer surface the batches depend on. Mirrors
// `queueAppService.push(queue, job, data, opts)` from TS.
type BatchPusher interface {
	Push(queueName, jobName string, data any, config *goqueues.JobConfig) (*goqueues.QueueJob, error)
}

// PaymentBatch mirrors ntx-apps/libs/capital/src/api/payment/payment.batch.ts.
// For each Plan in the supplied period_type, the batch pages through 100 rows
// at a time and enqueues `queue/apps/capital/payment` jobs that downstream
// consumers (Phase 3 worker) drain.
type PaymentBatch struct {
	queue      BatchPusher
	planEntity *capitalplan.PlanEntity
	periodType string
	period     string
	duration   time.Duration
}

// NewPaymentBatch constructs a batch for the supplied periodType. The period
// label is computed from the previous period's start (e.g. "01-2026" for
// monthly), matching TS payment.batch.ts:21-28.
func NewPaymentBatch(queue BatchPusher, planEntity *capitalplan.PlanEntity, periodType string) *PaymentBatch {
	now := time.Now().UTC()
	prev := subtractPeriod(now, periodType)
	from := startOfPeriod(prev, periodType)
	return &PaymentBatch{
		queue:      queue,
		planEntity: planEntity,
		periodType: periodType,
		period:     from.Format(periodFormat(periodType)),
		duration:   periodDuration(periodType),
	}
}

// Name returns "captial:payment:<period>" (sic — TS uses the typo).
func (b *PaymentBatch) Name() string { return "captial:payment:" + b.period }

// Limit returns the per-page size (matches TS limit() = 100).
func (b *PaymentBatch) Limit() int { return 100 }

// Length counts the eligible plans.
func (b *PaymentBatch) Length() (int64, error) {
	if b.planEntity == nil {
		return 0, nil
	}
	return b.planEntity.Count(`"period_type" = ?`, b.periodType)
}

// Page returns the (page,perPage) slice of plans ordered by created_at asc.
func (b *PaymentBatch) Page(page, perPage int) ([]capitalplan.Plan, error) {
	if b.planEntity == nil {
		return nil, nil
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage
	return b.planEntity.Find(offset, perPage, `"period_type" = ? ORDER BY "created_at" ASC`, b.periodType)
}

// Handler enqueues one Plan onto the `payment` queue. Returns the underlying
// queue push error (or nil) so the caller can collect failures.
func (b *PaymentBatch) Handler(page, index int, plan *capitalplan.Plan) error {
	if b.queue == nil || plan == nil {
		return nil
	}
	_, err := b.queue.Push("queue/apps/capital", "payment", map[string]any{
		"name":  b.Name(),
		"page":  page,
		"index": index,
		"plan":  plan,
	}, &goqueues.JobConfig{RetryLimit: 3, RetryDelay: 60 * 30 * 1000})
	return err
}

// Timeout returns 90% of the period duration. Used by the host scheduler to
// release any held lock just before the next tick fires.
func (b *PaymentBatch) Timeout() time.Duration {
	return time.Duration(float64(b.duration) * 0.9)
}

// Run iterates every page and dispatches a queue push per plan. Returns the
// number of plans enqueued + the first error encountered (if any).
func (b *PaymentBatch) Run() (int, error) {
	if b.planEntity == nil || b.queue == nil {
		return 0, nil
	}
	limit := b.Limit()
	total, err := b.Length()
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}
	count := 0
	pages := int((total + int64(limit) - 1) / int64(limit))
	for page := 1; page <= pages; page++ {
		plans, err := b.Page(page, limit)
		if err != nil {
			return count, err
		}
		for i := range plans {
			plan := plans[i]
			if err := b.Handler(page, i, &plan); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}

// RunFromQueueApp is a convenience wrapper that adapts queueapp.AppService —
// the gox-apps queue producer service — to the BatchPusher interface. Hides
// the goose generic queue indirection from callers in capital/app/app.service.go.
func RunFromQueueApp(qsvc *queueapp.AppService, planEntity *capitalplan.PlanEntity, periodType string) (int, error) {
	if qsvc == nil {
		return 0, nil
	}
	b := NewPaymentBatch(qsvc, planEntity, periodType)
	return b.Run()
}

// periodDuration mirrors TS convertPeriodDuration.
func periodDuration(periodType string) time.Duration {
	switch periodType {
	case PeriodDaily:
		return time.Duration(0.9*24*60*60) * time.Second
	case PeriodWeekly:
		return time.Duration(0.9*7*24*60*60) * time.Second
	case PeriodMonthly:
		return time.Duration(0.9*28*24*60*60) * time.Second
	case PeriodYearly:
		return time.Duration(0.9*365*24*60*60) * time.Second
	}
	return 0
}
