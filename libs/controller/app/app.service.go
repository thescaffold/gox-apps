package app

import (
	"github.com/thescaffold/gox-apps/libs/controller/app/route"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// RouteType mirrors TS RouteTypes enum in app.dto.ts.
const (
	RouteTypeHttp = "http"
	RouteTypeWs   = "ws"
)

// AppService mirrors ntx-apps/libs/controller/src/app.service.ts.
type AppService struct {
	routeEntity *route.RouteEntity `inject:""`
}

// GetHello returns the constant TS getHello() returns.
func (s *AppService) GetHello() string { return "Hello World!" }

// GetRoutes mirrors TS AppService.getRoutes(length). It is unused by the HTTP
// surface (TS doesn't expose it as a route either) but kept here so callers
// that wire it directly continue to work.
func (s *AppService) GetRoutes(limit int) ([]route.Route, error) {
	return s.routeEntity.Find(0, limit, "")
}

// GetRoute mirrors TS AppService.getRoute(): it derives the route components
// with extractPath, looks the route up by group+service, then appends
// resource / resourceId / component to the upstream when each is present.
func (s *AppService) GetRoute(path string) (string, error) {
	p := utils.ExtractPath(path)

	r, err := s.routeEntity.First(`"group" = ? AND "service" = ?`, p.Group, p.Service)
	if err != nil || r == nil {
		return "", nil
	}

	output := r.Upstream
	if p.Resource != "" {
		output += "/" + p.Resource
	}
	if p.ResourceID != "" {
		output += "/" + p.ResourceID
	}
	if p.Component != "" {
		output += "/" + p.Component
	}
	return output, nil
}

// RegisterRoutes mirrors TS AppController.register(): the dto payload is
// duplicated into a Http and a Ws route, each upserted by the
// (group, service, type, name) tuple. The update covers desc, upstream, status.
func (s *AppService) RegisterRoutes(dto *RegisterDto) error {
	for _, t := range []string{RouteTypeHttp, RouteTypeWs} {
		row := &route.Route{
			Group:    dto.Group,
			Service:  dto.Service,
			Type:     t,
			Name:     dto.Name,
			Desc:     dto.Desc,
			Upstream: dto.Upstream,
			Status:   dto.Status,
		}
		existing, err := s.routeEntity.First(
			`"group" = ? AND "service" = ? AND "type" = ? AND "name" = ?`,
			dto.Group, dto.Service, t, dto.Name,
		)
		if err != nil {
			return err
		}
		if existing != nil {
			if _, err := s.routeEntity.Update(row,
				`"group" = ? AND "service" = ? AND "type" = ? AND "name" = ?`,
				dto.Group, dto.Service, t, dto.Name,
			); err != nil {
				return err
			}
			continue
		}
		if err := s.routeEntity.Insert(row); err != nil {
			return err
		}
	}
	return nil
}
