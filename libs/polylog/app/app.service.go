package app

import (
	"encoding/json"

	polylogconfig "github.com/thescaffold/gox-apps-polylog/app/config"
	"github.com/thescaffold/gox-apps-polylog/pkg/pipeline"
	ntxctx "github.com/thescaffold/gox-packages-core/context"
)

// AppService surfaces polylog ingest/config operations.
// Mirrors ntx-apps/libs/polylog/src/app.service.ts.
type AppService struct {
	pipelineService *pipeline.PipelineService    `inject:""`
	configService   *polylogconfig.ConfigService `inject:""`
}

func (s *AppService) GetHello() string { return "Hello World!" }

// IngestBatch processes a batch of items.
// Mirrors TS AppService.ingestBatch().
func (s *AppService) IngestBatch(items []pipeline.Item, ctx ntxctx.NTXContext) ([]any, error) {
	out := make([]any, 0, len(items))
	for _, it := range items {
		ev, err := s.pipelineService.Ingest(it, ctx.UserID, ctx.ClientID, ctx.WorkspaceID)
		if err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, nil
}

// Ingest processes a single item. Mirrors TS AppService.ingest().
func (s *AppService) Ingest(item pipeline.Item, ctx ntxctx.NTXContext) (any, error) {
	return s.pipelineService.Ingest(item, ctx.UserID, ctx.ClientID, ctx.WorkspaceID)
}

// IngestSourceAlt is `POST /ingest/:id` — the :id is the source id; body+queries are merged.
func (s *AppService) IngestSourceAlt(body json.RawMessage, sourceID string, properties map[string]string, ctx ntxctx.NTXContext) (any, error) {
	item := pipeline.Item{Source: sourceID, Payload: body}
	if t, ok := properties["type"]; ok {
		item.Type = t
	}
	if c, ok := properties["category"]; ok {
		item.Category = c
	}
	return s.pipelineService.Ingest(item, ctx.UserID, ctx.ClientID, ctx.WorkspaceID)
}

// IngestSource is `POST /ingest/source/:id`. TS body sets entityName="source".
func (s *AppService) IngestSource(body json.RawMessage, sourceID string, properties map[string]string, ctx ntxctx.NTXContext) (any, error) {
	item := pipeline.Item{Source: sourceID, EntityName: "source", Payload: body}
	if t, ok := properties["type"]; ok {
		item.Type = t
	}
	if c, ok := properties["category"]; ok {
		item.Category = c
	}
	return s.pipelineService.Ingest(item, ctx.UserID, ctx.ClientID, ctx.WorkspaceID)
}

// IngestChannel is `POST /ingest/channel/:id`. TS body sets entityName="channel".
func (s *AppService) IngestChannel(body json.RawMessage, channelID string, properties map[string]string, ctx ntxctx.NTXContext) (any, error) {
	item := pipeline.Item{Source: channelID, EntityName: "channel", Payload: body}
	if t, ok := properties["type"]; ok {
		item.Type = t
	}
	if c, ok := properties["category"]; ok {
		item.Category = c
	}
	return s.pipelineService.Ingest(item, ctx.UserID, ctx.ClientID, ctx.WorkspaceID)
}

// configKeyFromBody pulls the identifier tuple out of a free-form body map.
// Tolerant of missing keys — empty string is fine for "not specified".
func configKeyFromBody(body map[string]any) (entityId, entityName, typ, category string) {
	if v, ok := body["entityId"].(string); ok {
		entityId = v
	}
	if v, ok := body["entityName"].(string); ok {
		entityName = v
	}
	if v, ok := body["type"].(string); ok {
		typ = v
	}
	if v, ok := body["category"].(string); ok {
		category = v
	}
	return
}

// SetConfig upserts a polylog config row by the (user, client, workspace,
// entityId, entityName, type, category) tuple. The "value" key in body is
// stored as the JSON value column. Mirrors TS updateOrCreate(tuple, {value}).
func (s *AppService) SetConfig(body map[string]any, ctx ntxctx.NTXContext) (any, error) {
	entityId, entityName, typ, category := configKeyFromBody(body)
	value, _ := json.Marshal(body["value"])
	var status *string
	if v, ok := body["status"].(string); ok {
		status = &v
	}
	return s.configService.Upsert(
		ctx.UserID, ctx.ClientID, ctx.WorkspaceID,
		entityId, entityName, typ, category,
		value, status,
	)
}

// GetConfig looks up a config row by the full tuple of identifiers.
// Returns nil when none exists, matching TS findOne semantics.
func (s *AppService) GetConfig(userID, clientID, workspaceID, entityID, entityName, typ, category string, ctx ntxctx.NTXContext) (any, error) {
	// Allow caller-supplied context to fill blanks.
	if userID == "" {
		userID = ctx.UserID
	}
	if clientID == "" {
		clientID = ctx.ClientID
	}
	if workspaceID == "" {
		workspaceID = ctx.WorkspaceID
	}
	row, err := s.configService.FindByTuple(userID, clientID, workspaceID, entityID, entityName, typ, category)
	if err != nil {
		return nil, err
	}
	return row, nil
}

// UpdateConfig patches an existing config row identified by id.
func (s *AppService) UpdateConfig(id string, body map[string]any, ctx ntxctx.NTXContext) (any, error) {
	value, _ := json.Marshal(body["value"])
	var status *string
	if v, ok := body["status"].(string); ok {
		status = &v
	}
	return s.configService.UpdateById(id, value, status)
}
