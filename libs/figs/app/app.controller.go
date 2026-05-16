package app

import (
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// AppController mirrors ntx-apps/libs/figs/src/app.controller.ts. Endpoints:
//
//	GET  /         getHello
//	POST /:name    saveFile (find-or-error + validate → map → convert → store)
type AppController struct {
	appService *AppService   `inject:""`
	lang       *i18n.Service `inject:""`
	log        types.Log     `inject:""`
}

// GetHello mirrors TS @Get() getHello(): success(getHello()) — the data is the
// raw string "Hello World!", no envelope title/message.
func (c *AppController) GetHello(dto *HelloDto) types.Output {
	return response.Success(c.appService.GetHello(), "", "", nil)
}

// SaveFile mirrors TS @Post(':name') saveFile. Existing-name → 'existing' error;
// any pipeline failure (validator/mapper/store) → 'invalid-request' error;
// success returns {url} + title/message via i18n.
func (c *AppController) SaveFile(dto *SaveFileDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.figs.app.title", nil, pref)

	name := dto.Name
	if name == "" {
		if n, ok := dto.MetaRaw["name"].(string); ok {
			name = n
		}
	}

	existing, _ := c.appService.FindFileByName(name)
	if existing != nil {
		return response.BadRequest(
			title,
			c.lang.Translate("apps.figs.app.post.file.error.existing", nil, pref),
		)
	}

	payload := &Payload{
		Meta:   dto.MetaRaw,
		Input:  dto.InputRaw,
		Output: dto.OutputRaw,
	}
	if payload.Meta == nil {
		payload.Meta = map[string]any{}
	}

	url, err := c.appService.SaveFile(name, payload)
	if err != nil || url == "" {
		return response.BadRequest(
			title,
			c.lang.Translate("apps.figs.app.post.file.error.invalid-request", nil, pref),
		)
	}

	return response.Success(map[string]any{"url": url},
		title,
		c.lang.Translate("apps.figs.app.post.file.success", nil, pref),
		nil)
}
