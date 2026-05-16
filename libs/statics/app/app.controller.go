package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// AppController mirrors ntx-apps/libs/statics/src/app.controller.ts. Endpoints:
//
//	GET  /                 getHello
//	GET  /filter/:key      filterByKey
//	GET  /code/:code       findOneByCode
//	GET  /value/:value     findOneByValue
type AppController struct {
	appService *AppService   `inject:""`
	lang       *i18n.Service `inject:""`
	log        types.Log     `inject:""`
}

// GetHello mirrors TS @Get() getHello(): success(getHello()) — plain string,
// no envelope title/message.
func (c *AppController) GetHello(dto *HelloDto) types.Output {
	return response.Success(c.appService.GetHello(), "", "", nil)
}

// FilterByKey mirrors TS @Get('filter/:key') filterByKey: returns success(list,
// title, msg). TS always returns success (with the list) even when empty.
func (c *AppController) FilterByKey(dto *FilterByKeyDto) types.Output {
	pref := dto.Ctx.Preference
	items, _ := c.appService.FilterByKey(dto.Key, dto.ParentId)
	return response.Success(items,
		c.lang.Translate("apps.statics.app.title", nil, pref),
		c.lang.Translate("apps.statics.app.get.filter.success", nil, pref),
		nil)
}

// FindByCode mirrors TS @Get('code/:code') findOneByCode: returns success(item,
// title, msg). TS always returns success (item is null when not found).
func (c *AppController) FindByCode(dto *FindByCodeDto) types.Output {
	pref := dto.Ctx.Preference
	item, _ := c.appService.FindByCode(dto.Code, dto.ParentId)
	return response.Success(item,
		c.lang.Translate("apps.statics.app.title", nil, pref),
		c.lang.Translate("apps.statics.app.get.filter.success", nil, pref),
		nil)
}

// FindByValue mirrors TS @Get('value/:value') findOneByValue: returns success(
// item, title, msg). TS always returns success.
func (c *AppController) FindByValue(dto *FindByValueDto) types.Output {
	pref := dto.Ctx.Preference
	item, _ := c.appService.FindByValue(dto.Value, dto.ParentId)
	return response.Success(item,
		c.lang.Translate("apps.statics.app.title", nil, pref),
		c.lang.Translate("apps.statics.app.get.filter.success", nil, pref),
		nil)
}
