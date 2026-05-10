package app

import (
	"strings"

	"github.com/awesome-goose/goose/io/output"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps-assets/app/file"
	"github.com/thescaffold/gox-packages-core/response"
)

type AppController struct {
	appService  *AppService       `inject:""`
	fileService *file.FileService `inject:""`
	log         types.Log         `inject:""`
}

func (c *AppController) Health(dto *HealthDto) types.Output {
	return response.Success(map[string]any{"status": c.appService.GetHello()}, "assets", "ok", nil)
}

func (c *AppController) DynamicFile(dto *DynamicFileDto) types.Output {
	// Fetch existing file by id
	if dto.Id != nil && *dto.Id != "" {
		f, err := c.fileService.FindById(*dto.Id)
		if err != nil || f == nil {
			return response.NotFound("assets", "File not found")
		}
		raw := ""
		if f.Raw != nil {
			raw = *f.Raw
		}
		return output.HTML(raw)
	}

	// Generate new SVG
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
		// TODO: inject ASSETS_BASE_URL via config when available.
		f, url, err := c.appService.StoreDynamicFile(name, svgContent, "")
		if err != nil {
			return response.InternalServerError("assets", "Failed to store dynamic file")
		}
		// Mirrors TS: returns file + computed URL.
		return response.Success(map[string]any{"file": f, "url": url}, "assets", "ok", nil)
	}

	// raw=true and the default both return the SVG body directly,
	// matching TS where the flag controls headers, not envelope.
	return output.HTML(svgContent)
}
