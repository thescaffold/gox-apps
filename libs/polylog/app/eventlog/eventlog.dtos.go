package eventlog

import (
	"encoding/json"

	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// DashboardDto carries the required `?type=` filter for GET /event-log/dashboard.
type DashboardDto struct {
	NTX  ntxctx.NTXContext `context:"ntx"`
	Type string            `query:"type"`
}

type CreateEventLogDto struct {
	EventId  string          `json:"eventId"  binding:"required"`
	Meta     json.RawMessage `json:"meta,omitempty"`
	Request  json.RawMessage `json:"request,omitempty"`
	Response json.RawMessage `json:"response,omitempty"`
	Status   *string         `json:"status,omitempty"`
}
type UpdateEventLogDto struct {
	Status *string `json:"status,omitempty"`
}
