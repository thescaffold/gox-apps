package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// AppController mirrors ntx-apps/libs/cache/src/app.controller.ts. The TS
// controller has no health endpoint and no /flush endpoint — both are absent
// here too. Every response is i18n-translated via `apps.cache.app.*` keys.
type AppController struct {
	appService *AppService   `inject:""`
	lang       *i18n.Service `inject:""`
	log        types.Log     `inject:""`
}

// Get mirrors TS @Get('get'). TS returns success(list, …) where list may be
// null — we deliberately do NOT 404 on a miss (TS returns 200 + data:null).
func (c *AppController) Get(dto *GetDto) types.Output {
	pref := dto.Ctx.Preference
	item, _ := c.appService.Get(dto.Key)
	return response.Success(item,
		c.lang.Translate("apps.cache.app.title", nil, pref),
		c.lang.Translate("apps.cache.app.get.success", nil, pref),
		nil)
}

// Set mirrors TS @Post('set'). The TS response data is the literal string 'OK'
// — `success('OK', …)` — independent of whether the row was inserted or updated.
func (c *AppController) Set(dto *SetDto) types.Output {
	pref := dto.Ctx.Preference
	if _, err := c.appService.Set(dto.Key, dto.Value, dto.Duration); err != nil {
		return response.InternalServerError(
			c.lang.Translate("apps.cache.app.title", nil, pref),
			err.Error(),
		)
	}
	return response.Success("OK",
		c.lang.Translate("apps.cache.app.title", nil, pref),
		c.lang.Translate("apps.cache.app.set.success", nil, pref),
		nil)
}

// SetNx mirrors TS @Post('setnx'). Returns 1 on insert, 0 if the key existed.
// TS uses the `apps.cache.app.setnx.success` translation key.
func (c *AppController) SetNx(dto *SetDto) types.Output {
	pref := dto.Ctx.Preference
	n, err := c.appService.SetNx(dto.Key, dto.Value, dto.Duration)
	if err != nil {
		return response.InternalServerError(
			c.lang.Translate("apps.cache.app.title", nil, pref),
			err.Error(),
		)
	}
	return response.Success(n,
		c.lang.Translate("apps.cache.app.title", nil, pref),
		c.lang.Translate("apps.cache.app.setnx.success", nil, pref),
		nil)
}

// GetSet mirrors TS @Post('getset'). Returns the PREVIOUS entry (or null when
// the key was unset). TS uses the `apps.cache.app.setnx.success` key here too
// (slightly odd but verbatim).
func (c *AppController) GetSet(dto *SetDto) types.Output {
	pref := dto.Ctx.Preference
	prev, err := c.appService.GetSet(dto.Key, dto.Value, dto.Duration)
	if err != nil {
		return response.InternalServerError(
			c.lang.Translate("apps.cache.app.title", nil, pref),
			err.Error(),
		)
	}
	return response.Success(prev,
		c.lang.Translate("apps.cache.app.title", nil, pref),
		c.lang.Translate("apps.cache.app.setnx.success", nil, pref),
		nil)
}

// Del mirrors TS @Delete('del'). Returns the number of rows deleted. TS
// uses the `apps.cache.app.set.success` translation key for this response.
func (c *AppController) Del(dto *DelDto) types.Output {
	pref := dto.Ctx.Preference
	n, err := c.appService.Del(dto.Key)
	if err != nil {
		return response.InternalServerError(
			c.lang.Translate("apps.cache.app.title", nil, pref),
			err.Error(),
		)
	}
	return response.Success(n,
		c.lang.Translate("apps.cache.app.title", nil, pref),
		c.lang.Translate("apps.cache.app.set.success", nil, pref),
		nil)
}

// TTL mirrors TS @Get('ttl'). -2 = missing, -1 = no expiry, >=0 = seconds.
// TS uses the `apps.cache.app.ttl.none` key for the -2 / -1 cases and the
// `apps.cache.app.ttl.success` key when an expiry is present.
func (c *AppController) TTL(dto *TTLDto) types.Output {
	pref := dto.Ctx.Preference
	secs, err := c.appService.TTL(dto.Key)
	if err != nil {
		return response.InternalServerError(
			c.lang.Translate("apps.cache.app.title", nil, pref),
			err.Error(),
		)
	}
	msgKey := "apps.cache.app.ttl.success"
	if secs < 0 {
		msgKey = "apps.cache.app.ttl.none"
	}
	return response.Success(secs,
		c.lang.Translate("apps.cache.app.title", nil, pref),
		c.lang.Translate(msgKey, nil, pref),
		nil)
}

// Incr mirrors TS @Post('incr'). Returns the post-increment integer, or 0 if
// the existing value couldn't be parsed.
func (c *AppController) Incr(dto *IncrDto) types.Output {
	pref := dto.Ctx.Preference
	n, err := c.appService.Incr(dto.Key)
	if err != nil {
		return response.InternalServerError(
			c.lang.Translate("apps.cache.app.title", nil, pref),
			err.Error(),
		)
	}
	return response.Success(n,
		c.lang.Translate("apps.cache.app.title", nil, pref),
		c.lang.Translate("apps.cache.app.incr.success", nil, pref),
		nil)
}
