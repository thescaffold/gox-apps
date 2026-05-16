// Package usage hosts the capital UsageService — emits `apps.capital.usage.*`
// events on the bus rather than persisting Usage rows directly. Mirrors
// ntx-apps/libs/capital/src/api/usage/usage.service.ts (the persistence is
// driven by the subscription handler on capital/app/app.service.go).
package usage

import (
	"time"

	"github.com/thescaffold/gox-packages/libs/core/events"
)

// UsageService publishes the four usage lifecycle events. Bridge calls
// UpdateUsage from its license-job consumer to register per-plan usage.
type UsageService struct {
	entity  *UsageEntity           `inject:""`
	tracker *events.TrackerService `inject:""`
}

// StartUsage emits apps.capital.usage.start.
func (s *UsageService) StartUsage(userID, clientID, workspaceID, name, id string, meta any, quantity int) {
	if s.tracker == nil {
		return
	}
	if quantity <= 0 {
		quantity = 1
	}
	s.tracker.Message("apps.capital.usage.start", map[string]any{
		"userId":      userID,
		"clientId":    clientID,
		"workspaceId": workspaceID,
		"entityName":  name,
		"entityId":    id,
		"quantity":    quantity,
		"startAt":     time.Now().UTC(),
		"meta":        meta,
	})
}

// UpdateUsage emits apps.capital.usage.update.
func (s *UsageService) UpdateUsage(userID, clientID, workspaceID, name, id string, meta any, quantity int) {
	if s.tracker == nil {
		return
	}
	if quantity <= 0 {
		quantity = 1
	}
	s.tracker.Message("apps.capital.usage.update", map[string]any{
		"userId":      userID,
		"clientId":    clientID,
		"workspaceId": workspaceID,
		"entityName":  name,
		"entityId":    id,
		"quantity":    quantity,
		"meta":        meta,
	})
}

// StopUsage emits apps.capital.usage.stop.
func (s *UsageService) StopUsage(userID, clientID, workspaceID, name, id string, meta any) {
	if s.tracker == nil {
		return
	}
	s.tracker.Message("apps.capital.usage.stop", map[string]any{
		"userId":      userID,
		"clientId":    clientID,
		"workspaceId": workspaceID,
		"entityName":  name,
		"entityId":    id,
		"stopAt":      time.Now().UTC(),
		"meta":        meta,
	})
}
