package app

import (
	"strings"

	"github.com/thescaffold/gox-apps-controller/app/route"
)

type AppService struct {
	routeEntity *route.RouteEntity `inject:""`
}

func (s *AppService) GetHello() string { return "Hello World!" }

func (s *AppService) GetRoutes(limit int) ([]route.Route, error) {
	return s.routeEntity.Find(0, limit, "")
}

func (s *AppService) GetRoute(path string) (string, error) {
	parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 5)
	if len(parts) < 2 {
		return "", nil
	}
	group, service := parts[0], parts[1]

	r, err := s.routeEntity.First(`"group" = ? AND "service" = ?`, group, service)
	if err != nil || r == nil {
		return "", nil
	}

	upstream := r.Upstream
	for _, part := range parts[2:] {
		upstream += "/" + part
	}
	return upstream, nil
}
