package task

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/awesome-goose/goose/types"
	healthlog "github.com/thescaffold/gox-apps/libs/health/app/log"
	healthsvc "github.com/thescaffold/gox-apps/libs/health/app/service"
	healthsum "github.com/thescaffold/gox-apps/libs/health/app/summary"
)

type SummaryTask struct {
	serviceEntity *healthsvc.ServiceEntity `inject:""`
	logEntity     *healthlog.LogEntity     `inject:""`
	summaryEntity *healthsum.SummaryEntity `inject:""`
}

// Boot starts the summary cron goroutine — runs every 10 minutes.
func (t *SummaryTask) Boot(_ types.Kernel) error {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			t.run()
		}
	}()
	return nil
}

func (t *SummaryTask) run() {
	toggle := os.Getenv("SUMMARY_TOGGLE")
	if toggle != "" && toggle != "on" {
		return
	}

	intervalSecs := 3600
	if v := os.Getenv("SUMMARY_INTERVAL_SECS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			intervalSecs = n
		}
	}

	expectedCount := 12
	if v := os.Getenv("SUMMARY_EXPECTED_EVENT_COUNT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			expectedCount = n
		}
	}

	from := time.Now().UTC().Add(-time.Duration(intervalSecs) * time.Second)
	fromStr := from.Format("2006-01-02 15:04:05")

	services, err := t.serviceEntity.All()
	if err != nil {
		return
	}

	for _, svc := range services {
		count, err := t.logEntity.Count(
			fmt.Sprintf(`"service_id" = ? AND "created_at" >= '%s'`, fromStr),
			svc.Id,
		)
		if err != nil {
			continue
		}

		received := int(count)
		// TS's leftJoin + WHERE log.createdAt >= from excludes services with no
		// logs in the window — they get no summary and keep their prior type. Skip
		// them here too (this also makes the measure<=0 "down" branch unreachable,
		// exactly as in TS).
		if received == 0 {
			continue
		}
		measure := 0.0
		if expectedCount > 0 {
			measure = float64(received) / float64(expectedCount)
		}

		outcomeType := outcomeFromMeasure(measure)

		// TS save({...}) with no id INSERTs a new summary every run (history
		// accumulates); it does not update a single per-service row.
		_ = t.summaryEntity.Insert(&healthsum.Summary{
			ServiceId: svc.Id,
			Type:      outcomeType,
			Received:  received,
			Measure:   measure,
		})

		updatedSvc := &healthsvc.Service{Type: &outcomeType}
		_, _ = t.serviceEntity.Update(updatedSvc, `"id" = ?`, svc.Id)
	}
}

func outcomeFromMeasure(measure float64) string {
	switch {
	case measure <= 0:
		return "down"
	case measure < 0.5:
		return "troubled"
	case measure < 0.8:
		// TS HealthStatusType.PossibleTroubled = 'possibly_troubled'.
		return "possibly_troubled"
	default:
		return "good"
	}
}
