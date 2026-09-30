package store

import (
	"github.com/thescaffold/gox-apps/libs/figs/pkg/converter"
	"github.com/thescaffold/gox-packages/libs/blobs/objectstore"
)

type ProviderType string

const (
	Local  ProviderType = "local"
	Object ProviderType = "object"
)

type Response struct {
	URL string
	Raw string
}

type Payload struct {
	Meta   map[string]any `json:"meta"`
	Input  map[string]any `json:"input"`
	Output map[string]any `json:"output"`
}

type Provider interface {
	Store(payload *Payload, raw *converter.Response) (*Response, error)
}

// Service hands out providers. Objects is the ObjectStore the `object`
// provider writes to (the control plane injects the store selected by
// BLOBS_BACKEND — Postgres by default).
type Service struct {
	Objects objectstore.ObjectStore
}

func (s *Service) Use(t ProviderType) Provider {
	switch t {
	case Object:
		return &ObjectProvider{Objects: s.Objects}
	default:
		return &LocalProvider{}
	}
}
