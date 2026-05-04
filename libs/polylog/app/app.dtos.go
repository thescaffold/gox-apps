package app

import (
	"encoding/json"

	ntxctx "github.com/thescaffold/gox-packages-core/context"
)

type HealthDto struct{}

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
