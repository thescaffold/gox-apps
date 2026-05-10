package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps-polylog/pkg/pipeline"
	"github.com/thescaffold/gox-packages-core/response"
)

type AppController struct {
	appService      *AppService               `inject:""`
	pipelineService *pipeline.PipelineService `inject:""`
}

// Health is GET /.
func (c *AppController) Health(dto *HealthDto) types.Output {
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "polylog", "ok", nil)
}

// Ingest handles POST /ingest.
func (c *AppController) Ingest(dto *IngestDto) types.Output {
	item := pipeline.Item{
		EntityId:   dto.EntityId,
		EntityName: dto.EntityName,
		Source:     dto.Source,
		Category:   dto.Category,
		Type:       dto.Type,
		Reference:  dto.Reference,
		Version:    dto.Version,
		Payload:    dto.Payload,
	}
	ev, err := c.appService.Ingest(item, dto.NTX)
	if err != nil {
		return response.InternalServerError("polylog", err.Error())
	}
	return response.Success(ev, "polylog", "event ingested", nil)
}

// IngestBatch handles POST /ingest/batch.
func (c *AppController) IngestBatch(dto *IngestBatchDto) types.Output {
	out, err := c.appService.IngestBatch(dto.Items, dto.NTX)
	if err != nil {
		return response.InternalServerError("polylog", err.Error())
	}
	return response.Success(out, "polylog", "batch ingested", nil)
}

// IngestSourceAlt handles POST /ingest/:id.
func (c *AppController) IngestSourceAlt(dto *IngestSourceParamDto) types.Output {
	ev, err := c.appService.IngestSourceAlt(dto.Payload, dto.ID, dto.Properties, dto.NTX)
	if err != nil {
		return response.InternalServerError("polylog", err.Error())
	}
	return response.Success(ev, "polylog", "event ingested", nil)
}

// IngestSource handles POST /ingest/source/:id.
func (c *AppController) IngestSource(dto *IngestSourceParamDto) types.Output {
	ev, err := c.appService.IngestSource(dto.Payload, dto.ID, dto.Properties, dto.NTX)
	if err != nil {
		return response.InternalServerError("polylog", err.Error())
	}
	return response.Success(ev, "polylog", "event ingested", nil)
}

// IngestChannel handles POST /ingest/channel/:id.
func (c *AppController) IngestChannel(dto *IngestSourceParamDto) types.Output {
	ev, err := c.appService.IngestChannel(dto.Payload, dto.ID, dto.Properties, dto.NTX)
	if err != nil {
		return response.InternalServerError("polylog", err.Error())
	}
	return response.Success(ev, "polylog", "event ingested", nil)
}

// SetConfig handles POST /config.
func (c *AppController) SetConfig(dto *SetConfigDto) types.Output {
	out, err := c.appService.SetConfig(dto.Body, dto.NTX)
	if err != nil {
		return response.InternalServerError("polylog", err.Error())
	}
	return response.Success(out, "polylog", "ok", nil)
}

// GetConfig handles GET /config.
func (c *AppController) GetConfig(dto *GetConfigDto) types.Output {
	out, err := c.appService.GetConfig(
		dto.UserId, dto.ClientId, dto.WorkspaceId,
		dto.EntityId, dto.EntityName, dto.Type, dto.Category,
		dto.NTX,
	)
	if err != nil {
		return response.InternalServerError("polylog", err.Error())
	}
	return response.Success(out, "polylog", "ok", nil)
}

// UpdateConfig handles PATCH /config/:id.
func (c *AppController) UpdateConfig(dto *UpdateConfigDto) types.Output {
	out, err := c.appService.UpdateConfig(dto.ID, dto.Body, dto.NTX)
	if err != nil {
		return response.InternalServerError("polylog", err.Error())
	}
	return response.Success(out, "polylog", "ok", nil)
}
