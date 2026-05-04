package validator

type ProviderType string

const (
	JSON ProviderType = "json"
	YAML ProviderType = "yaml"
)

type Payload struct {
	Meta   map[string]any `json:"meta"`
	Input  map[string]any `json:"input"`
	Output map[string]any `json:"output"`
}

type Provider interface {
	Validate(payload *Payload) (bool, error)
}

type Service struct{}

func (s *Service) Use(t ProviderType) Provider {
	switch t {
	case YAML:
		return &YAMLProvider{}
	default:
		return &JSONProvider{}
	}
}
