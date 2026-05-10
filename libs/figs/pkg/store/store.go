package store

import "github.com/thescaffold/gox-apps/libs/figs/pkg/converter"

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

type Service struct{}

func (s *Service) Use(t ProviderType) Provider {
	switch t {
	case Object:
		return &ObjectProvider{}
	default:
		return &LocalProvider{}
	}
}
