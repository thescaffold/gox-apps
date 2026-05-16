package app

import (
	goqueues "github.com/awesome-goose/goose/modules/queues"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// AppController mirrors ntx-apps/libs/queue/src/app.controller.ts. Three
// endpoints registered under `apps/queue`: POST / (push), GET / (pop),
// PATCH / (log). Every response title/message is i18n-translated.
type AppController struct {
	appService *AppService   `inject:""`
	lang       *i18n.Service `inject:""`
}

// Push mirrors TS @Post() push(). On a nil/erroring result TS raises an
// error() envelope — error() defaults to BAD_REQUEST.
func (c *AppController) Push(dto *PushDto) types.Output {
	pref := dto.Ctx.Preference

	data, err := c.appService.Push(dto.Queue, dto.Job, dto.Data, dto.Config.toJobConfig())
	if err != nil || data == nil {
		return response.BadRequest(
			c.lang.Translate("apps.queue.app.title", nil, pref),
			c.lang.Translate("apps.queue.app.push.error.invalid-request",
				map[string]any{"queue": dto.Queue, "job": dto.Job}, pref),
		)
	}
	return response.Success(data,
		c.lang.Translate("apps.queue.app.title", nil, pref),
		c.lang.Translate("apps.queue.app.push.success", nil, pref),
		nil)
}

// Pop mirrors TS @Get() pop(). Missing queue/job returns a translated error
// envelope (with only the `queue` interpolation key, per TS); a missing job
// returns null in data (NOT an error, per the TS comment).
func (c *AppController) Pop(dto *PopDto) types.Output {
	pref := dto.Ctx.Preference

	if dto.Queue == "" || dto.Job == "" {
		return response.BadRequest(
			c.lang.Translate("apps.queue.app.title", nil, pref),
			c.lang.Translate("apps.queue.app.pop.error.invalid-request",
				map[string]any{"queue": dto.Queue}, pref),
		)
	}

	data, err := c.appService.Pop(dto.Queue, dto.Job)
	if err != nil {
		return response.InternalServerError(
			c.lang.Translate("apps.queue.app.title", nil, pref),
			err.Error(),
		)
	}
	return response.Success(data,
		c.lang.Translate("apps.queue.app.title", nil, pref),
		c.lang.Translate("apps.queue.app.pop.success", nil, pref),
		nil)
}

// Log mirrors TS @Patch() log(). An invalid status or unknown jobId causes the
// appService to return nil, surfaced as a translated error envelope.
func (c *AppController) Log(dto *LogDto) types.Output {
	pref := dto.Ctx.Preference

	data, err := c.appService.Log(dto.JobId, dto.Status, dto.Output)
	if err != nil || data == nil {
		return response.BadRequest(
			c.lang.Translate("apps.queue.app.title", nil, pref),
			c.lang.Translate("apps.queue.app.log.error.invalid-request", nil, pref),
		)
	}
	return response.Success(data,
		c.lang.Translate("apps.queue.app.title", nil, pref),
		c.lang.Translate("apps.queue.app.log.success", nil, pref),
		nil)
}

// toJobConfig converts the wire-level PushConfigDto into the underlying goose
// JobConfig. A nil receiver, or a payload with every optional field absent,
// produces a nil config — matching TS which only forwards `config` to the
// appService when at least one inner field was supplied.
func (pc *PushConfigDto) toJobConfig() *goqueues.JobConfig {
	if pc == nil {
		return nil
	}
	out := &goqueues.JobConfig{}
	hasField := false
	if pc.Priority != nil {
		out.Priority = *pc.Priority
		hasField = true
	}
	if pc.RetryLimit != nil {
		out.RetryLimit = *pc.RetryLimit
		hasField = true
	}
	if pc.RetryDelay != nil {
		out.RetryDelay = *pc.RetryDelay
		hasField = true
	}
	if pc.StartAt != nil {
		out.StartAt = pc.StartAt
		hasField = true
	}
	if pc.ExpireAt != nil {
		out.ExpireAt = pc.ExpireAt
		hasField = true
	}
	if pc.Singleton != nil {
		out.Singleton = *pc.Singleton
		hasField = true
	}
	if pc.Frequency != nil {
		out.Frequency = *pc.Frequency
		hasField = true
	}
	if !hasField {
		return nil
	}
	return out
}
