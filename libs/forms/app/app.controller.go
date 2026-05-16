package app

import (
	"errors"

	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// AppController mirrors ntx-apps/libs/forms/src/app.controller.ts.
//
// Note: TS imports `error` from 'node:console' instead of the ntx-core helper,
// which means TS getForm/saveForm "errors" are silently dropped (the handler
// returns undefined and Nest serialises that to an empty 200 body). That is
// almost certainly a TS bug; gox returns the intended translated error envelope
// (400, the BAD_REQUEST default of ntx-core error()) so clients get a usable
// signal — flip to response.Success(nil, …) if strict parity is ever required.
type AppController struct {
	appService *AppService   `inject:""`
	lang       *i18n.Service `inject:""`
}

// GetHello mirrors TS @Get() getHello(): success(this.appService.getHello()).
func (c *AppController) GetHello(dto *HelloDto) types.Output {
	return response.Success(c.appService.GetHello(), "", "", nil)
}

// GetForm mirrors TS @Get('one/:id') getForm().
func (c *AppController) GetForm(dto *GetFormDto) types.Output {
	pref := dto.Ctx.Preference

	view, err := c.appService.GetForm(dto.Id)
	if err != nil {
		return response.BadRequest(
			c.lang.Translate("apps.forms.app.title", nil, pref),
			c.lang.Translate("apps.forms.app.error.not-found", nil, pref),
		)
	}
	return response.Success(view,
		c.lang.Translate("apps.forms.app.title", nil, pref),
		c.lang.Translate("apps.forms.app.success.form-retrieved", nil, pref),
		nil)
}

// SaveForm mirrors TS @Post('one/:id') saveForm(). A missing field returns
// `apps.forms.app.error.field-not-found` with {key} interpolation.
func (c *AppController) SaveForm(dto *SaveFormDto) types.Output {
	pref := dto.Ctx.Preference

	err := c.appService.SaveForm(dto.Id, dto.Body)
	if err != nil {
		title := c.lang.Translate("apps.forms.app.title", nil, pref)
		var missing *FieldNotFoundError
		if errors.As(err, &missing) {
			return response.BadRequest(
				title,
				c.lang.Translate(
					"apps.forms.app.error.field-not-found",
					map[string]any{"key": missing.Key},
					pref,
				),
			)
		}
		if errors.Is(err, ErrFormNotFound) {
			return response.BadRequest(
				title,
				c.lang.Translate("apps.forms.app.error.not-found", nil, pref),
			)
		}
		return response.InternalServerError(title, err.Error())
	}
	return response.Success(nil,
		c.lang.Translate("apps.forms.app.title", nil, pref),
		c.lang.Translate("apps.forms.app.success.form-saved", nil, pref),
		nil)
}
