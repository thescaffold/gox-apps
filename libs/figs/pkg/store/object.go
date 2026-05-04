package store

import "github.com/thescaffold/gox-apps-figs/pkg/converter"

type ObjectProvider struct{}

func (p *ObjectProvider) Store(payload *Payload, raw *converter.Response) (*Response, error) {
	// stub: object storage (S3/GCS) not yet implemented
	return &Response{URL: "", Raw: ""}, nil
}
