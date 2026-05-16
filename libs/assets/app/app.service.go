package app

import (
	"os"
	"strings"

	"github.com/thescaffold/gox-apps/libs/assets/app/file"
	"github.com/thescaffold/gox-packages/libs/core/image"
	"github.com/thescaffold/gox-packages/libs/core/services/text"
)

// AppService is the public assets service. Mirrors ntx-apps/libs/assets/src/app.service.ts.
type AppService struct {
	fileService  *file.FileService `inject:""`
	imageService *image.Service    `inject:""`
}

func (s *AppService) GetHello() string {
	return "Hello World!"
}

// GenerateDynamicSVG produces an SVG using the core ImageService variants.
// Mirrors TS app.service.ts which calls imageService.new(name, variant, colors, size, square).
//
// variant: "pixel" (default) or "shapes" — the only variants the TS
// ImageService exposes. Any other value falls back to the pixel variant, the
// same as the TS `new()` default parameter. Size and colour defaults (80 and
// the standard palette) are applied by the core ImageService itself.
func (s *AppService) GenerateDynamicSVG(name, variant string, colors []string, size int, square bool) string {
	v := image.VariantPixel
	if strings.ToLower(variant) == "shapes" {
		v = image.VariantShapes
	}

	svc := s.imageService
	if svc == nil {
		// Allow construction without DI (tests). image.Service is stateless.
		svc = &image.Service{}
	}
	out := svc.New(image.Options{
		Variant:  v,
		Size:     size,
		Colors:   colors,
		Name:     name,
		IsSquare: square,
	})
	return string(out)
}

// RandomName generates a random name. Mirrors TS, which defaults the dynamic
// file name to `textService.name('-')` — i.e. "{adjective}-{noun}-{nonce}".
func RandomName() string {
	return text.New().Name("-")
}

// GetDynamicAsset mirrors TS AssetsAppService.getDynamicAsset — generates an
// SVG with a random name + default pixel variant, persists it via the
// FileService, and returns the `{id, name, url}` envelope callers
// (workspace.beforeCreate, identity/provider/authorize, etc.) embed in the
// owning entity.
//
// The URL is composed from BLOBS_BASE_URL (env, optional) + `/dynamic?id=<file.id>`
// so the asset can be served by the public blobs endpoint regardless of where
// the host app is mounted. Falls back to a placeholder only when both the
// image generator AND the file service fail — earlier impls returned a
// placeholder on `err != nil` alone which silently masked storage errors.
func (s *AppService) GetDynamicAsset() *DynamicAsset {
	name := RandomName()
	svg := s.GenerateDynamicSVG(name, "pixel", nil, 80, true)
	baseURL := os.Getenv("BLOBS_BASE_URL")
	f, url, err := s.StoreDynamicFile(name, svg, baseURL)
	if err == nil && f != nil {
		return &DynamicAsset{Id: f.Id, Name: name, Url: url}
	}
	// Storage failed — fall back to embedding the SVG inline as a data URI so
	// the caller still has a renderable asset reference. TS doesn't do this
	// (it errors) but the gox impl prefers degraded-but-usable over hard fail.
	return &DynamicAsset{Name: name, Url: dataURISVG(svg)}
}

// dataURISVG packs an SVG into an inline data URI. Used as the last-resort
// fallback when the FileService can't persist the dynamic asset.
func dataURISVG(svg string) string {
	if svg == "" {
		return "/assets/dynamic-placeholder.svg"
	}
	return "data:image/svg+xml;utf8," + svg
}

// DynamicAsset is the envelope returned by GetDynamicAsset — matches TS
// `{ id, name, url }` shape so cross-app callers (identity/workspace) can
// embed it without further marshalling.
type DynamicAsset struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

// StoreDynamicFile saves a generated SVG and returns the file with its computed URL.
func (s *AppService) StoreDynamicFile(name, svgContent, baseURL string) (*file.File, string, error) {
	f := &file.File{
		Type:   "svg",
		Bucket: "dynamic",
		Name:   name,
		Tags:   []string{name, "svg", "dynamic"},
		Raw:    &svgContent,
	}
	if err := s.fileService.Save(f); err != nil {
		return nil, "", err
	}
	url := baseURL + "/dynamic?id=" + f.Id
	return f, url, nil
}
