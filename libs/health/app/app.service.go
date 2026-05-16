package app

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"

	healthlog "github.com/thescaffold/gox-apps/libs/health/app/log"
	healthsvc "github.com/thescaffold/gox-apps/libs/health/app/service"
	healthsum "github.com/thescaffold/gox-apps/libs/health/app/summary"
)

// AppService backs the health app endpoints and event subscriptions. Mirrors
// ntx-apps/libs/health/src/app.controller.ts + app.service.ts.
type AppService struct {
	logService     *healthlog.LogService     `inject:""`
	serviceService *healthsvc.ServiceService `inject:""`
	logEntity      *healthlog.LogEntity      `inject:""`
	summaryEntity  *healthsum.SummaryEntity  `inject:""`
}

// GetHello mirrors TS AppService.getHello().
func (s *AppService) GetHello() string { return "Hello World!" }

// Ping handles POST /ping logic: find the service by name, write a log entry,
// then update the service's state column. Mirrors TS AppController.ping().
func (s *AppService) Ping(name, state string, meta json.RawMessage) error {
	svc, err := s.serviceService.FindByName(name)
	if err != nil || svc == nil {
		return errors.New("service not found")
	}
	if _, err := s.logService.CreateWithMeta(svc.Id, state, meta); err != nil {
		return err
	}
	return s.serviceService.UpdateState(svc.Id, state)
}

// Register handles POST /register logic: upsert a Service row by name.
// Mirrors TS AppController.register(): only desc + status reach
// updateOrCreate (state/type are server-managed).
func (s *AppService) Register(name string, desc, status *string) (*healthsvc.Service, error) {
	return s.serviceService.UpsertByName(name, desc, status)
}

// Housekeep mirrors TS apps.cron.heartbeat.hourly subscription: it deletes
// Health/Summary and Health/Log rows older than LOG_RETENTION_THRESHOLD hours
// (default 24). gox publishes the cleanup itself instead of relying on the
// scheduler; the subscription handler in index.go calls this method.
func (s *AppService) Housekeep() error {
	threshold := 24
	if v := os.Getenv("LOG_RETENTION_THRESHOLD"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			threshold = n
		}
	}
	cutoff := time.Now().UTC().Add(-time.Duration(threshold) * time.Hour)
	cutoffStr := cutoff.Format("2006-01-02 15:04:05")

	if _, err := s.summaryEntity.Delete(`"created_at" < ?`, cutoffStr); err != nil {
		return err
	}
	if _, err := s.logEntity.Delete(`"created_at" < ?`, cutoffStr); err != nil {
		return err
	}
	return nil
}
