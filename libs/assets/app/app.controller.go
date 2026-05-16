package app

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/awesome-goose/goose/io/output"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/assets/app/file"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// AppController mirrors ntx-apps/libs/assets/src/app.controller.ts.
type AppController struct {
	appService  *AppService       `inject:""`
	fileService *file.FileService `inject:""`
	lang        *i18n.Service     `inject:""`
	log         types.Log         `inject:""`
}

// svgHeaders mirrors the headers the TS dynamicFile endpoint attaches to a
// rendered SVG response: `{ 'content-type': 'image/svg+xml', 'Cache-Control':
// 'max-age=0, s-maxage=2592000' }`.
var svgHeaders = map[string]string{
	"Cache-Control": "max-age=0, s-maxage=2592000",
}

// GetHello mirrors TS @Get() getHello(): success(this.appService.getHello()).
// TS calls success() with data only — the response has no title/message.
func (c *AppController) GetHello(dto *HelloDto) types.Output {
	return response.Success(c.appService.GetHello(), "", "", nil)
}

// DynamicFile mirrors TS AppController.dynamicFile().
//
//   - id set: look the stored file up; on miss return a 400 error envelope
//     (TS error() defaults to BAD_REQUEST) with a translated message, on hit
//     stream file.raw as SVG with the cache headers.
//   - otherwise: render a new SVG. The name defaults to textService.name('-').
//     When store=true the file is persisted and returned flat ({...file, url})
//     with translated title/message. When store is false the SVG body is
//     streamed directly — with the SVG content-type + cache headers unless
//     raw=true (TS sends raw with no header overrides).
func (c *AppController) DynamicFile(dto *DynamicFileDto) types.Output {
	pref := dto.Ctx.Preference

	// Fetch existing file by id.
	if dto.Id != nil && *dto.Id != "" {
		f, err := c.fileService.FindById(*dto.Id)
		if err != nil || f == nil {
			return response.BadRequest(
				c.lang.Translate("apps.assets.app.title", nil, pref),
				c.lang.Translate("apps.assets.app.get.dynamic.error.not-found",
					map[string]any{"id": *dto.Id}, pref),
			)
		}
		raw := ""
		if f.Raw != nil {
			raw = *f.Raw
		}
		return output.Raw(raw,
			output.WithRawContentType("image/svg+xml"),
			output.WithRawHeaders(svgHeaders),
		)
	}

	// Generate a new SVG.
	name := RandomName()
	if dto.Name != nil && *dto.Name != "" {
		name = *dto.Name
	}

	variant := "pixel"
	if dto.Variant != nil {
		variant = *dto.Variant
	}

	var colors []string
	if dto.Colors != nil && *dto.Colors != "" {
		colors = strings.Split(*dto.Colors, ",")
	}

	svgContent := c.appService.GenerateDynamicSVG(name, variant, colors, dto.Size, dto.Square)

	if dto.Store {
		// Mirrors TS `${configService.get('ASSETS_BASE_URL')}/dynamic?id=...`.
		baseURL := os.Getenv("ASSETS_BASE_URL")
		f, url, err := c.appService.StoreDynamicFile(name, svgContent, baseURL)
		if err != nil {
			return response.InternalServerError(
				c.lang.Translate("apps.assets.app.title", nil, pref),
				c.lang.Translate("apps.assets.app.get.dynamic.error.store-failed", nil, pref),
			)
		}
		// TS returns success({ ...file, url }, title, message) — file fields
		// flattened with url, plus the standard translated title/message.
		return response.Success(flattenFileWithURL(f, url),
			c.lang.Translate("apps.assets.app.title", nil, pref),
			c.lang.Translate("apps.assets.app.get.dynamic.success", nil, pref),
			nil)
	}

	// raw=true streams the SVG body with no header overrides (TS passes
	// headers=null); otherwise the SVG content-type + cache headers apply.
	if dto.Raw {
		return output.Raw(svgContent)
	}
	return output.Raw(svgContent,
		output.WithRawContentType("image/svg+xml"),
		output.WithRawHeaders(svgHeaders),
	)
}

// flattenFileWithURL converts a File into a flat map and adds the url key,
// mirroring the TS `{ ...file, url }` spread.
func flattenFileWithURL(f *file.File, url string) map[string]any {
	out := map[string]any{}
	if b, err := json.Marshal(f); err == nil {
		_ = json.Unmarshal(b, &out)
	}
	out["url"] = url
	return out
}
