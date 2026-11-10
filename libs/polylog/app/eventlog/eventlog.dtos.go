package eventlog

import (
	"encoding/json"
	"github.com/awesome-goose/goose/openapi"

	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// DashboardDto carries the required `?type=` filter for GET /event-log/dashboard.
type DashboardDto struct {
	NTX  ntxctx.NTXContext `context:"ntx"`
	Type string            `query:"type"`
}

// Dashboard is what GET /event-log/dashboard answers: platform-wide counts, for a platform admin.
type Dashboard struct {
	Total      int64 `json:"total" binding:"required"`
	Completed  int64 `json:"completed" binding:"required"`
	InProgress int64 `json:"inProgress" binding:"required"`
}

func (DashboardDto) OpenAPI() openapi.Operation {
	return openapi.Operation{Summary: "Counts of completed and in-progress event-logs (platform admin)", Response: Dashboard{}}
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
