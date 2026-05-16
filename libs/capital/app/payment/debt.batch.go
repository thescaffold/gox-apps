package payment

import (
	"strings"
	"time"

	goqueues "github.com/awesome-goose/goose/modules/queues"
	capitalplan "github.com/thescaffold/gox-apps/libs/capital/app/plan"
	queueapp "github.com/thescaffold/gox-apps/libs/queue/app"
)

// DebtBatch mirrors ntx-apps/libs/capital/src/api/payment/debt.batch.ts.
// Same shape as PaymentBatch but filters to plans with unpaid recurrent
// Payment rows, and pushes onto the `debt` job instead of `payment`.
type DebtBatch struct {
	queue      BatchPusher
	planEntity *capitalplan.PlanEntity
	entity     *PaymentEntity
	periodType string
	period     string
	duration   time.Duration
}

// NewDebtBatch constructs a batch for the supplied periodType. period is
// computed from the previous period's start (TS-equivalent).
func NewDebtBatch(queue BatchPusher, planEntity *capitalplan.PlanEntity, entity *PaymentEntity, periodType string) *DebtBatch {
	now := time.Now().UTC()
	prev := subtractPeriod(now, periodType)
	from := startOfPeriod(prev, periodType)
	return &DebtBatch{
		queue:      queue,
		planEntity: planEntity,
		entity:     entity,
		periodType: periodType,
		period:     from.Format(periodFormat(periodType)),
		duration:   periodDuration(periodType),
	}
}

// Name returns "capital:debt:<period>".
func (b *DebtBatch) Name() string { return "capital:debt:" + b.period }

// Limit returns 100.
func (b *DebtBatch) Limit() int { return 100 }

// recurrentTypesInClause builds the SQL `IN (...)` placeholders + args for
// PaymentMetaRecurrentTypes (the TS-equivalent allowlist).
func recurrentTypesInClause() (string, []any) {
	types := []string{
		"apps.capital.payment.renewal",
		"apps.capital.record.pay",
	}
	placeholders := make([]string, len(types))
	args := make([]any, 0, len(types))
	for i, t := range types {
		placeholders[i] = "?"
		args = append(args, t)
	}
	return "(" + strings.Join(placeholders, ",") + ")", args
}

// Length counts plans with at least one unpaid recurrent payment.
func (b *DebtBatch) Length() (int64, error) {
	if b.planEntity == nil || b.entity == nil {
		return 0, nil
	}
	inClause, args := recurrentTypesInClause()
	args = append([]any{b.periodType}, args...)
	args = append(args, false)
	// EXISTS subquery: plans whose id appears in any unpaid recurrent Payment.
	where := `"period_type" = ? AND "id" IN (SELECT "plan_id" FROM "CapitalPayments" WHERE "type" IN ` + inClause + ` AND ("paid" IS NULL OR "paid" = ?))`
	return b.planEntity.Count(where, args...)
}

// Page returns the page of plans with unpaid recurrent payments.
func (b *DebtBatch) Page(page, perPage int) ([]capitalplan.Plan, error) {
	if b.planEntity == nil {
		return nil, nil
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage
	inClause, args := recurrentTypesInClause()
	args = append([]any{b.periodType}, args...)
	args = append(args, false)
	where := `"period_type" = ? AND "id" IN (SELECT "plan_id" FROM "CapitalPayments" WHERE "type" IN ` + inClause + ` AND ("paid" IS NULL OR "paid" = ?)) ORDER BY "created_at" ASC`
	return b.planEntity.Find(offset, perPage, where, args...)
}

// Handler enqueues a debt job for one plan.
func (b *DebtBatch) Handler(page, index int, plan *capitalplan.Plan) error {
	if b.queue == nil || plan == nil {
		return nil
	}
	_, err := b.queue.Push("queue/apps/capital", "debt", map[string]any{
		"name":  b.Name(),
		"page":  page,
		"index": index,
		"plan":  plan,
	}, &goqueues.JobConfig{RetryLimit: 3, RetryDelay: 60 * 30 * 1000})
	return err
}

// Timeout returns 90% of the period duration.
func (b *DebtBatch) Timeout() time.Duration {
	return time.Duration(float64(b.duration) * 0.9)
}

// Run drains every page and enqueues a debt job per matching plan.
func (b *DebtBatch) Run() (int, error) {
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

// RunDebtFromQueueApp adapts queueapp.AppService for the DebtBatch path.
func RunDebtFromQueueApp(qsvc *queueapp.AppService, planEntity *capitalplan.PlanEntity, entity *PaymentEntity, periodType string) (int, error) {
	if qsvc == nil {
		return 0, nil
	}
	b := NewDebtBatch(qsvc, planEntity, entity, periodType)
	return b.Run()
}
