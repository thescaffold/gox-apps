package app

import (
	"encoding/json"

	"github.com/thescaffold/gox-apps/libs/polylog/pkg/pipeline"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

type HealthDto struct{}

// IngestDto carries one ingest item.
type IngestDto struct {
	NTX        ntxctx.NTXContext `context:"ntx"`
	EntityId   string            `json:"entityId"`
	EntityName string            `json:"entityName"`
	Source     string            `json:"source"`
	Category   string            `json:"category"`
	Type       string            `json:"type"`
	Reference  string            `json:"reference"`
	Version    string            `json:"version"`
	Payload    json.RawMessage   `json:"payload"`
}

// IngestBatchDto mirrors TS Items wrapper.
type IngestBatchDto struct {
	NTX   ntxctx.NTXContext `context:"ntx"`
	Items []pipeline.Item   `json:"items"`
}

// IngestSourceParamDto carries the :id path param + arbitrary query map for
// /ingest/:id, /ingest/source/:id and /ingest/channel/:id.
type IngestSourceParamDto struct {
	NTX        ntxctx.NTXContext `context:"ntx"`
	ID         string            `param:"id"`
	Properties map[string]string `query:"-"` // populated by handler from raw queries
	Payload    json.RawMessage   `json:"-"`  // populated by handler from raw body
}

// SetConfigDto mirrors TS SetConfigDto — flexible map body.
type SetConfigDto struct {
	NTX  ntxctx.NTXContext `context:"ntx"`
	Body map[string]any    `json:",inline"`
}

// GetConfigDto carries the lookup query params for GET /config.
type GetConfigDto struct {
	NTX         ntxctx.NTXContext `context:"ntx"`
	UserId      string            `query:"userId"`
	ClientId    string            `query:"clientId"`
	WorkspaceId string            `query:"workspaceId"`
	EntityId    string            `query:"entityId"`
	EntityName  string            `query:"entityName"`
	Type        string            `query:"type"`
	Category    string            `query:"category"`
}

// UpdateConfigDto carries the :id path param + patch body for PATCH /config/:id.
type UpdateConfigDto struct {
	NTX  ntxctx.NTXContext `context:"ntx"`
	ID   string            `param:"id"`
	Body map[string]any    `json:",inline"`
}
