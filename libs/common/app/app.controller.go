package app

import (
	"encoding/json"
	"strings"

	"github.com/awesome-goose/goose/types"
	staticsapp "github.com/thescaffold/gox-apps/libs/statics/app"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// AppController mirrors ntx-apps/libs/common/src/app.controller.ts. Endpoints:
//
//	GET  /                      getHello
//	GET  /currency/:currency    getCurrency
//	GET  /location              getLocation
type AppController struct {
	appService *AppService           `inject:""`
	staticsApp *staticsapp.AppService `inject:""`
	lang       *i18n.Service         `inject:""`
}

// GetHello mirrors TS @Get() getHello(): success(getHello()) — plain string,
// no envelope title/message.
func (c *AppController) GetHello(dto *HelloDto) types.Output {
	return response.Success(c.appService.GetHello(), "", "", nil)
}

// GetCurrency mirrors TS @Get('currency/:currency'): success(rate, title, msg).
// TS does not 404 on a missing currency — it returns success with the raw
// repository result (null when absent), so we mirror that behaviour exactly.
func (c *AppController) GetCurrency(dto *GetCurrencyDto) types.Output {
	pref := dto.Ctx.Preference
	rate, _ := c.appService.GetCurrency(dto.Currency)
	return response.Success(rate,
		c.lang.Translate("apps.common.title", nil, pref),
		c.lang.Translate("apps.common.get.currency.success", nil, pref),
		nil)
}

// GetLocation mirrors TS @Get('location') — find-or-create the Ip row, look
// up the country and timezone via StaticsAppService, then return the
// {theme, timezone, country, currency, language} envelope TS produces.
// When the country can't be resolved, TS returns an error envelope keyed
// `apps.capital.app.location.error.no-country` (note: TS uses the capital
// namespace here — preserved verbatim for surface parity).
func (c *AppController) GetLocation(dto *GetLocationDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.common.title", nil, pref)

	ipValue := dto.Ip
	if ipValue == "" {
		ipValue = dto.HeaderIP
	}
	if ipValue == "" {
		return response.BadRequest(title, "ip is required")
	}

	loc, err := c.appService.GetLocation(ipValue)
	if err != nil || loc == nil {
		return response.InternalServerError(title, "location lookup failed")
	}

	// TS: const country = await this.staticsAppService.filterByCode(ip.country);
	country, _ := c.staticsApp.FindByCode(loc.Country, nil)
	if country == nil {
		return response.BadRequest(
			c.lang.Translate("apps.capital.app.title", nil, pref),
			c.lang.Translate("apps.capital.app.location.error.no-country", nil, pref),
		)
	}

	// country.meta carries { emoji, currency: {code,name,symbol}, language: {code,name} }.
	type currencyMeta struct {
		Code   string `json:"code"`
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
	}
	type languageMeta struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
	var meta struct {
		Emoji    string       `json:"emoji"`
		Currency currencyMeta `json:"currency"`
		Language languageMeta `json:"language"`
	}
	if len(country.Meta) > 0 {
		_ = json.Unmarshal(country.Meta, &meta)
	}

	// TS: filterByKey('timezone', ip.country) → take first .value.
	parent := loc.Country
	timezone := "America/New_York"
	if zones, _ := c.staticsApp.FilterByKey("timezone", &parent); len(zones) > 0 {
		if zones[0].Value != "" {
			timezone = zones[0].Value
		}
	}

	code := ""
	if country.Code != nil {
		code = *country.Code
	}

	data := map[string]any{
		"theme":    "light",
		"timezone": timezone,
		"country": map[string]any{
			"code": strings.ToUpper(code),
			"name": country.Value,
			"icon": meta.Emoji,
		},
		"currency": map[string]any{
			"code": strings.ToUpper(meta.Currency.Code),
			"name": meta.Currency.Name,
			"icon": meta.Currency.Symbol,
		},
		"language": map[string]any{
			"code": strings.ToUpper(meta.Language.Code),
			"name": meta.Language.Name,
			"icon": nil,
		},
	}

	return response.Success(data,
		title,
		c.lang.Translate("apps.common.get.location.success", nil, pref),
		nil)
}
