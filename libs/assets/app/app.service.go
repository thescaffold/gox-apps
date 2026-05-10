package app

import (
	"strings"

	"github.com/thescaffold/gox-apps/libs/assets/app/file"
	"github.com/thescaffold/gox-packages/libs/core/image"
	"github.com/thescaffold/gox-packages/libs/core/utils"
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
// variant: "pixel" (default), "shapes", "gradient", or "solid".
// When colors is empty, falls back to the same default palette TS uses.
func (s *AppService) GenerateDynamicSVG(name, variant string, colors []string, size int, square bool) string {
	if size <= 0 {
		size = 80
	}
	if len(colors) == 0 {
		colors = []string{"#92A1C6", "#146A7C", "#F0AB3D", "#C271B4", "#C20D90"}
	}

	v := image.VariantPixel
	switch strings.ToLower(variant) {
	case "shapes":
		v = image.VariantShapes
	case "gradient":
		v = image.VariantGradient
	case "solid":
		v = image.VariantSolid
	}

	svc := s.imageService
	if svc == nil {
		// Allow construction without DI (tests). image.Service is stateless.
		svc = &image.Service{}
	}
	out := svc.New(image.Options{
		Variant:  v,
		Width:    size,
		Height:   size,
		Colors:   colors,
		Name:     name,
		IsSquare: square,
	})
	return string(out)
}

// RandomName generates a random name (mirrors TS TextService.username()).
func RandomName() string {
	return utils.RandomUsername()
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
