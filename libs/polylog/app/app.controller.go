package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/polylog/pkg/pipeline"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// AppController mirrors ntx-apps/libs/polylog/src/app.controller.ts. Endpoints:
//
//	GET    /                   getHello
//	POST   /ingest/batch       ingestBatch
//	POST   /ingest             ingest
//	POST   /ingest/:id         ingestSourceAlt
//	POST   /ingest/source/:id  ingestSource
//	POST   /ingest/channel/:id ingestChannel
//	POST   /config             setConfig
//	GET    /config             getConfig
//	PATCH  /config/:id         updateConfig
//	POST   /root-source        upsertRootSource
//	POST   /root-sink          upsertRootSink
type AppController struct {
	appService      *AppService               `inject:""`
	pipelineService *pipeline.PipelineService `inject:""`
	lang            *i18n.Service             `inject:""`
}

// GetHello mirrors TS @Get() getHello(): success(getHello()) — plain string,
// no envelope title/message.
func (c *AppController) GetHello(dto *HelloDto) types.Output {
	return response.Success(c.appService.GetHello(), "", "", nil)
}

// Ingest handles POST /ingest.
func (c *AppController) Ingest(dto *IngestDto) types.Output {
	pref := dto.NTX.Preference
	title := c.lang.Translate("apps.polylog.app.title", nil, pref)
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
		return response.InternalServerError(title, err.Error())
	}
	return response.Success(ev,
		title,
		c.lang.Translate("apps.polylog.app.post.ingest.success", nil, pref),
		nil)
}

// IngestBatch handles POST /ingest/batch.
func (c *AppController) IngestBatch(dto *IngestBatchDto) types.Output {
	pref := dto.NTX.Preference
	title := c.lang.Translate("apps.polylog.app.title", nil, pref)
	out, err := c.appService.IngestBatch(dto.Items, dto.NTX)
	if err != nil {
		return response.InternalServerError(title, err.Error())
	}
	return response.Success(out,
		title,
		c.lang.Translate("apps.polylog.app.post.ingest.batch.success", nil, pref),
		nil)
}

// IngestSourceAlt handles POST /ingest/:id.
func (c *AppController) IngestSourceAlt(dto *IngestSourceParamDto) types.Output {
	pref := dto.NTX.Preference
	title := c.lang.Translate("apps.polylog.app.title", nil, pref)
	ev, err := c.appService.IngestSourceAlt(dto.Payload, dto.ID, dto.Properties, dto.NTX)
	if err != nil {
		return response.InternalServerError(title, err.Error())
	}
	return response.Success(ev,
		title,
		c.lang.Translate("apps.polylog.app.post.ingest.success", nil, pref),
		nil)
}

// IngestSource handles POST /ingest/source/:id.
func (c *AppController) IngestSource(dto *IngestSourceParamDto) types.Output {
	pref := dto.NTX.Preference
	title := c.lang.Translate("apps.polylog.app.title", nil, pref)
	ev, err := c.appService.IngestSource(dto.Payload, dto.ID, dto.Properties, dto.NTX)
	if err != nil {
		return response.InternalServerError(title, err.Error())
	}
	return response.Success(ev,
		title,
		c.lang.Translate("apps.polylog.app.post.ingest.source.success", nil, pref),
		nil)
}

// IngestChannel handles POST /ingest/channel/:id.
func (c *AppController) IngestChannel(dto *IngestSourceParamDto) types.Output {
	pref := dto.NTX.Preference
	title := c.lang.Translate("apps.polylog.app.title", nil, pref)
	ev, err := c.appService.IngestChannel(dto.Payload, dto.ID, dto.Properties, dto.NTX)
	if err != nil {
		return response.InternalServerError(title, err.Error())
	}
	return response.Success(ev,
		title,
		c.lang.Translate("apps.polylog.app.post.ingest.channel.success", nil, pref),
		nil)
}

// SetConfig handles POST /config — the 3-way source/sink/channel upsert.
func (c *AppController) SetConfig(dto *SetConfigDto) types.Output {
	pref := dto.NTX.Preference
	title := c.lang.Translate("apps.polylog.app.title", nil, pref)
	out, err := c.appService.SetConfig(dto)
	if err != nil {
		return response.InternalServerError(title, err.Error())
	}
	return response.Success(out,
		title,
		c.lang.Translate("apps.polylog.app.post.config.success", nil, pref),
		nil)
}

// GetConfig handles GET /config.
func (c *AppController) GetConfig(dto *GetConfigDto) types.Output {
	pref := dto.NTX.Preference
	title := c.lang.Translate("apps.polylog.app.title", nil, pref)
	out, err := c.appService.GetConfig(
		dto.UserId, dto.ClientId, dto.WorkspaceId,
		dto.EntityId, dto.EntityName, dto.Type, dto.Category,
		dto.NTX,
	)
	if err != nil {
		return response.InternalServerError(title, err.Error())
	}
	return response.Success(out,
		title,
		c.lang.Translate("apps.polylog.app.get.config.success", nil, pref),
		nil)
}

// UpdateConfig handles PATCH /config/:id. TS returns an error envelope when
// the underlying source doesn't exist; gox mirrors that.
func (c *AppController) UpdateConfig(dto *UpdateConfigDto) types.Output {
	pref := dto.NTX.Preference
	title := c.lang.Translate("apps.polylog.app.title", nil, pref)
	out, err := c.appService.UpdateConfig(dto.ID, dto.Status, dto.NTX)
	if err != nil {
		return response.InternalServerError(title, err.Error())
	}
	if out == nil {
		return response.BadRequest(title,
			c.lang.Translate("apps.polylog.app.patch.config.error.invalid-request", nil, pref),
		)
	}
	return response.Success(out,
		title,
		c.lang.Translate("apps.polylog.app.patch.config.success", nil, pref),
		nil)
}

// UpsertRootSource handles POST /root-source. TS swallows errors via .catch
// and returns success(undefined); gox mirrors that by ignoring the err.
func (c *AppController) UpsertRootSource(dto *UpsertRootSourceDto) types.Output {
	pref := dto.NTX.Preference
	out, _ := c.appService.CreateRootSource(dto)
	return response.Success(out,
		c.lang.Translate("apps.polylog.app.title", nil, pref),
		c.lang.Translate("apps.polylog.app.post.root-source.success", nil, pref),
		nil)
}

// UpsertRootSink handles POST /root-sink. Same swallow-error semantics as TS.
func (c *AppController) UpsertRootSink(dto *UpsertRootSinkDto) types.Output {
	pref := dto.NTX.Preference
	out, _ := c.appService.CreateRootSink(dto)
	return response.Success(out,
		c.lang.Translate("apps.polylog.app.title", nil, pref),
		c.lang.Translate("apps.polylog.app.post.root-sink.success", nil, pref),
		nil)
}
