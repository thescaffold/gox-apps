package app

import (
	gocron "github.com/awesome-goose/goose/modules/cron"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// AppController mirrors ntx-apps/libs/cron/src/app.controller.ts. Three
// endpoints registered under `apps/cron`: POST / (register), GET / (select),
// PATCH / (log). Every response title/message is i18n-translated.
type AppController struct {
	appService *AppService   `inject:""`
	lang       *i18n.Service `inject:""`
}

// Register mirrors TS @Post() register(). When the appService returns nil or
// errors, TS raises a translated error() envelope — error() defaults to
// BAD_REQUEST, so the gox 400 response matches.
func (c *AppController) Register(dto *RegisterDto) types.Output {
	pref := dto.Ctx.Preference

	data, err := c.appService.Register(dto.Group, dto.Name, dto.Pattern, dto.Config.toCronConfig())
	if err != nil || data == nil {
		return response.BadRequest(
			c.lang.Translate("apps.cron.app.title", nil, pref),
			c.lang.Translate("apps.cron.app.register.error.invalid-request",
				map[string]any{"group": dto.Group, "name": dto.Name}, pref),
		)
	}
	return response.Success(data,
		c.lang.Translate("apps.cron.app.title", nil, pref),
		c.lang.Translate("apps.cron.app.register.success", nil, pref),
		nil)
}

// Select mirrors TS @Get() select(). Missing group/name returns a translated
// error envelope; a missing job returns null in the data field (NOT an error,
// per the TS comment "return null if no job is available — don't throw").
func (c *AppController) Select(dto *SelectDto) types.Output {
	pref := dto.Ctx.Preference

	if dto.Group == "" || dto.Name == "" {
		return response.BadRequest(
			c.lang.Translate("apps.cron.app.title", nil, pref),
			c.lang.Translate("apps.cron.app.pop.error.invalid-request",
				map[string]any{"group": dto.Group, "name": dto.Name}, pref),
		)
	}

	data, err := c.appService.Select(dto.Group, dto.Name)
	if err != nil {
		return response.InternalServerError(
			c.lang.Translate("apps.cron.app.title", nil, pref),
			err.Error(),
		)
	}
	return response.Success(data,
		c.lang.Translate("apps.cron.app.title", nil, pref),
		c.lang.Translate("apps.cron.app.pop.success", nil, pref),
		nil)
}

// Log mirrors TS @Patch() log(). Invalid statuses or unknown jobIds cause the
// appService to return nil, which we surface as a translated error envelope.
func (c *AppController) Log(dto *LogDto) types.Output {
	pref := dto.Ctx.Preference

	data, err := c.appService.Log(dto.JobId, dto.Status, dto.Output)
	if err != nil || data == nil {
		return response.BadRequest(
			c.lang.Translate("apps.cron.app.title", nil, pref),
			c.lang.Translate("apps.cron.app.log.error.invalid-request", nil, pref),
		)
	}
	return response.Success(data,
		c.lang.Translate("apps.cron.app.title", nil, pref),
		c.lang.Translate("apps.cron.app.log.success", nil, pref),
		nil)
}

// toCronConfig translates the wire-level config DTO into the underlying goose
// cron config struct. A nil receiver and a payload with every optional field
// absent both produce a nil config — matching TS, which only forwards `config`
// when at least one inner field is supplied.
func (rc *RegisterConfigDto) toCronConfig() *gocron.CronConfig {
	if rc == nil {
		return nil
	}
	out := &gocron.CronConfig{}
	hasField := false
	if rc.Priority != nil {
		out.Priority = *rc.Priority
		hasField = true
	}
	if rc.RetryLimit != nil {
		out.RetryLimit = *rc.RetryLimit
		hasField = true
	}
	if rc.RetryDelay != nil {
		out.RetryDelay = *rc.RetryDelay
		hasField = true
	}
	if rc.StartAt != nil {
		out.StartAt = rc.StartAt
		hasField = true
	}
	if rc.ExpireAt != nil {
		out.ExpireAt = rc.ExpireAt
		hasField = true
	}
	if !hasField {
		return nil
	}
	return out
}
