package app

import (
	gocron "github.com/awesome-goose/goose/modules/cron"
	"github.com/thescaffold/gox-packages-core/events"
)

type AppService struct {
	cronSvc *gocron.Cron `inject:""`
	bus     *events.Bus  `inject:""`
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
