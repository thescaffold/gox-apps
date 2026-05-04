package app

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/thescaffold/gox-apps-assets/app/file"
)

type AppService struct {
	fileService *file.FileService `inject:""`
}

func (s *AppService) GetHello() string {
	return "Hello from assets"
}

// GenerateDynamicSVG produces a simple placeholder SVG.
// The TS version used ntx-core's ImageService; this is a faithful stub.
func (s *AppService) GenerateDynamicSVG(name, variant string, colors []string, size int, square bool) string {
	if size <= 0 {
		size = 100
	}
	color := "#6366f1"
	if len(colors) > 0 {
		color = colors[0]
	}
	initial := strings.ToUpper(name[:1])
	width := size
	height := size
	if !square {
		height = size * 2 / 3
	}
	_ = variant // variant affects style in TS; stub ignores it
	return fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`+
			`<rect width="%d" height="%d" fill="%s"/>`+
			`<text x="50%%" y="50%%" dominant-baseline="middle" text-anchor="middle" `+
			`font-family="sans-serif" font-size="%d" fill="#fff">%s</text>`+
			`</svg>`,
		width, height, width, height,
		width, height, color,
		size/3, initial,
	)
}

// RandomName generates a random hyphen-separated name (mirrors TS TextService.name).
func RandomName() string {
	words := []string{"swift", "calm", "bright", "bold", "keen", "vast", "pure", "deep"}
	return words[rand.Intn(len(words))] + "-" + words[rand.Intn(len(words))]
}

// StoreDynamicFile saves a generated SVG as a file record and returns it with its URL.
func (s *AppService) StoreDynamicFile(name, svgContent, baseURL string) (*file.File, error) {
	tags, _ := json.Marshal([]string{name, "svg", "dynamic"})
	f := &file.File{
		Type:   "svg",
		Bucket: "dynamic",
		Name:   name,
		Tags:   tags,
		Raw:    &svgContent,
	}
	if err := s.fileService.Save(f); err != nil {
		return nil, err
	}
	url := baseURL + "/dynamic?id=" + f.Id
	_ = url
	return f, nil
}
