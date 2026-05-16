package event

import (
	"encoding/json"

	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// DashboardDto carries the optional `?type=` filter for GET /event/dashboard.
type DashboardDto struct {
	NTX  ntxctx.NTXContext `context:"ntx"`
	Type string            `query:"type"`
}

// CreateEventDto mirrors ntx-apps/libs/polylog/src/api/event/dto/create-event.dto.ts.
// userId/clientId/workspaceId are NOT in the TS DTO — they are populated by
// the controller morph from request context.
type CreateEventDto struct {
	UserId      string          `json:"-"`
	ClientId    string          `json:"-"`
	WorkspaceId string          `json:"-"`
	EntityId    string          `json:"entityId"    binding:"required"`
	EntityName  string          `json:"entityName"  binding:"required"`
	Category    string          `json:"category"    binding:"required"`
	Reference   string          `json:"reference"   binding:"required"`
	Version     string          `json:"version"     binding:"required"`
	Payload     json.RawMessage `json:"payload"     binding:"required"`
	Type        *string         `json:"type,omitempty"`
	Status      *string         `json:"status,omitempty"`
}

type UpdateEventDto struct {
	Status *string `json:"status,omitempty"`
}
