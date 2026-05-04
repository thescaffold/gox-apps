package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/thescaffold/gox-apps-figs/pkg/converter"
)

type LocalProvider struct{}

func (p *LocalProvider) Store(payload *Payload, raw *converter.Response) (*Response, error) {
	name, _ := payload.Meta["name"].(string)
	bucket, _ := payload.Meta["bucket"].(string)
	if bucket == "" {
		bucket = "common"
	}
	dir := filepath.Join("storage", bucket)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	filename := fmt.Sprintf("%s.%s", name, raw.Extension)
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, raw.Buffer, 0644); err != nil {
		return nil, err
	}
	return &Response{
		URL: "/" + filepath.ToSlash(path),
		Raw: string(raw.Buffer),
	}, nil
}
