package batch

import (
	"fmt"
	"os"
	"time"

	"github.com/awesome-goose/goose/types"
	auditlog "github.com/thescaffold/gox-apps-audit/app/log"
)

// defaultBatch is set by BatchModule.Declarations so Trigger can call it after DI.
var defaultBatch *SummaryBatch

// Trigger is the events.EventHandler called from the apps.cron.heartbeat.hourly subscription.
func Trigger(_ string, _ any) {
	if defaultBatch != nil {
		go defaultBatch.run()
	}
}

// SummaryBatch runs an hourly aggregation of audit log activity.
// Boots a self-starting ticker; also callable via Trigger for event-driven firing.
type SummaryBatch struct {
	logEntity *auditlog.LogEntity `inject:""`
	logger    types.Log           `inject:""`
}

func (b *SummaryBatch) Boot(_ types.Kernel) error {
	defaultBatch = b
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			b.run()
		}
	}()
	return nil
}

func (b *SummaryBatch) run() {
	toggle := os.Getenv("AUDIT_BATCH_TOGGLE")
	if toggle != "" && toggle != "on" {
		return
	}

	intervalSecs := 3600
	from := time.Now().UTC().Add(-time.Duration(intervalSecs) * time.Second)
	fromStr := from.Format("2006-01-02 15:04:05")

	total, err := b.logEntity.Count(fmt.Sprintf(`"created_at" >= '%s'`, fromStr))
	if err != nil {
		return
	}

	described, err := b.logEntity.Count(fmt.Sprintf(`"created_at" >= '%s' AND "desc" IS NOT NULL`, fromStr))
	if err != nil {
		return
	}

	services := b.countByField("service", fromStr)
	entities := b.countByField("entity_name", fromStr)
	actions := b.countByField("action", fromStr)

	b.logger.Info(fmt.Sprintf(
		"[audit/batch] hourly summary: total=%d described=%d services=%v entities=%v actions=%v",
		total, described, services, entities, actions,
	))
}

// countByField counts log entries grouped by a set of known values for a column.
func (b *SummaryBatch) countByField(field, fromStr string) map[string]int64 {
	known := []string{"create", "update", "delete", "read"}
	counts := map[string]int64{}
	for _, val := range known {
		c, err := b.logEntity.Count(
			fmt.Sprintf(`"created_at" >= '%s' AND "%s" = ?`, fromStr, field),
			val,
		)
		if err == nil && c > 0 {
			counts[val] = c
		}
	}
	return counts
}
