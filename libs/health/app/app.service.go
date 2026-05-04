package app

import (
	healthlog "github.com/thescaffold/gox-apps-health/app/log"
	healthsvc "github.com/thescaffold/gox-apps-health/app/service"
)

type AppService struct {
	logService     *healthlog.LogService     `inject:""`
	serviceService *healthsvc.ServiceService `inject:""`
}

func (s *AppService) GetHello() string { return "Hello from health" }

func (s *AppService) Check(serviceId string) (*healthlog.Log, error) {
	svc, err := s.serviceService.FindById(serviceId)
	if err != nil || svc == nil {
		return nil, err
	}
	state := "ok"
	return s.logService.Create(serviceId, &state)
}
