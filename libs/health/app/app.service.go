package app

import (
	"encoding/json"
	"errors"

	healthlog "github.com/thescaffold/gox-apps-health/app/log"
	healthsvc "github.com/thescaffold/gox-apps-health/app/service"
)

type AppService struct {
	logService     *healthlog.LogService     `inject:""`
	serviceService *healthsvc.ServiceService `inject:""`
}

// GetHello mirrors TS AppService.getHello().
func (s *AppService) GetHello() string { return "Hello World!" }

// Check writes a synthetic "ok" log entry for the given service id.
func (s *AppService) Check(serviceId string) (*healthlog.Log, error) {
	svc, err := s.serviceService.FindById(serviceId)
	if err != nil || svc == nil {
		return nil, err
	}
	state := "ok"
	return s.logService.Create(serviceId, &state)
}

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
// Mirrors TS AppController.register().
func (s *AppService) Register(name string, desc, status *string) (*healthsvc.Service, error) {
	return s.serviceService.UpsertByName(name, desc, status)
}
