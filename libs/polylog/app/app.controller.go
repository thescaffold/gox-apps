package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps-polylog/pkg/pipeline"
	"github.com/thescaffold/gox-packages-core/response"
)

type AppController struct {
	appService      *AppService              `inject:""`
	pipelineService *pipeline.PipelineService `inject:""`
}

func (c *AppController) Health(dto *HealthDto) types.Output {
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "polylog", "ok", nil)
}

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
	ev, err := c.pipelineService.Ingest(item, dto.NTX.UserID, dto.NTX.ClientID, dto.NTX.WorkspaceID)
	if err != nil {
		return response.InternalServerError("polylog", err.Error())
	}
	return response.Success(ev, "polylog", "event ingested", nil)
}
