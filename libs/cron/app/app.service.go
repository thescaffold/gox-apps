package app

import (
	"time"

	gocron "github.com/awesome-goose/goose/modules/cron"
	cronlog "github.com/thescaffold/gox-apps/libs/cron/app/log"
	"github.com/thescaffold/gox-packages/libs/core/events"
)

type AppService struct {
	cronSvc *gocron.Cron        `inject:""`
	bus     *events.Bus         `inject:""`
	logs    *cronlog.LogEntity  `inject:""`
}

func (s *AppService) GetHello() string { return "Hello World!" }

func (s *AppService) Register(group, name, pattern string, config *gocron.CronConfig) (*gocron.CronJob, error) {
	return s.cronSvc.Register(group, name, pattern, config)
}

func (s *AppService) Select(group, name string) (*gocron.CronJob, error) {
	return s.cronSvc.Select(group, name)
}

func (s *AppService) Log(jobId, status string, output any) (*gocron.CronLog, error) {
	return s.cronSvc.Log(jobId, status, output)
}

func (s *AppService) PublishHeartbeat(eventType string, job *gocron.CronJob) {
	if s.bus != nil {
		s.bus.Publish(eventType, job)
	}
}

// Housekeep deletes successful CronLog rows older than cutoff.
// Mirrors TS AppController.subscriptions['apps.cron.heartbeat.hourly']
// which deletes Log rows where status=Success AND createdAt < threshold.
func (s *AppService) Housekeep(cutoff time.Time) {
	if s.logs == nil {
		return
	}
	_, _ = s.logs.Delete(`"status" = ? AND "created_at" < ?`, gocron.LogStatusSuccess, cutoff)
}
