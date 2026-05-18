package app

import (
	"time"

	"github.com/thescaffold/gox-apps/libs/controller/app/request"
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
	routeEntity   *route.RouteEntity     `inject:""`
	requestEntity *request.RequestEntity `inject:""`
}

// Housekeep deletes ControllerRequests rows older than cutoff. Mirrors TS
// AppController.subscriptions['apps.cron.heartbeat.hourly'] which deletes
// Request rows where createdAt < threshold (no status filter).
func (s *AppService) Housekeep(cutoff time.Time) {
	if s.requestEntity == nil {
		return
	}
	_, _ = s.requestEntity.Delete(`"created_at" < ?`, cutoff)
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
	return s.upsertRoutePair(dto.Group, dto.Service, dto.Name, dto.Upstream, dto.Desc, dto.Status)
}

// RegisterRoutePayload mirrors TS subscription
// 'apps.controller.route.register' which receives a raw payload object and
// upserts the same Http/Ws pair as the HTTP @Post('register') endpoint. The
// payload keys map to RegisterDto fields one-for-one; unrecognized keys are
// ignored. The function accepts the goose Item.Payload (any) shape directly.
func (s *AppService) RegisterRoutePayload(payload any) error {
	m, ok := payload.(map[string]any)
	if !ok {
		return nil
	}
	get := func(k string) string {
		if v, ok := m[k].(string); ok {
			return v
		}
		return ""
	}
	getPtr := func(k string) *string {
		if v, ok := m[k].(string); ok {
			return &v
		}
		return nil
	}
	return s.upsertRoutePair(get("group"), get("service"), get("name"), get("upstream"), getPtr("desc"), getPtr("status"))
}

// upsertRoutePair inserts or updates two Route rows (Http + Ws variants) keyed
// by (group, service, type, name) — the shared kernel of register / subscription.
// Returns nil when entity is unwired or the (group, service, name) tuple is
// incomplete (so empty subscription payloads don't insert garbage rows).
func (s *AppService) upsertRoutePair(group, service, name, upstream string, desc, status *string) error {
	if s.routeEntity == nil {
		return nil
	}
	if group == "" || service == "" || name == "" {
		return nil
	}
	for _, t := range []string{RouteTypeHttp, RouteTypeWs} {
		row := &route.Route{
			Group:    group,
			Service:  service,
			Type:     t,
			Name:     name,
			Desc:     desc,
			Upstream: upstream,
			Status:   status,
		}
		existing, err := s.routeEntity.First(
			`"group" = ? AND "service" = ? AND "type" = ? AND "name" = ?`,
			group, service, t, name,
		)
		if err != nil {
			return err
		}
		if existing != nil {
			if _, err := s.routeEntity.Update(row,
				`"group" = ? AND "service" = ? AND "type" = ? AND "name" = ?`,
				group, service, t, name,
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
