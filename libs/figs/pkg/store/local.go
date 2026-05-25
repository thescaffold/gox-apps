package store

import (
	"encoding/base64"

	"github.com/thescaffold/gox-apps/libs/figs/pkg/converter"
)

type LocalProvider struct{}

// Store mirrors TS LocalService.store: it does NOT write to disk — it returns
// the bare meta name as the url and the converted content as base64. The base64
// raw is persisted to file.raw; the object/S3 store handles real remote storage.
func (p *LocalProvider) Store(payload *Payload, raw *converter.Response) (*Response, error) {
	name, _ := payload.Meta["name"].(string)
	return &Response{
		URL: name,
		Raw: base64.StdEncoding.EncodeToString(raw.Buffer),
	}, nil
}
