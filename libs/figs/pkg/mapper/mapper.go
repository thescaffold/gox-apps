package mapper

type ProviderType string

const (
	HTML ProviderType = "html"
	Data ProviderType = "data"
)

type Payload struct {
	Meta   map[string]any `json:"meta"`
	Input  map[string]any `json:"input"`
	Output map[string]any `json:"output"`
}

type Provider interface {
	Map(payload *Payload) ([]any, error)
}

type Service struct{}

func (s *Service) Use(t ProviderType) Provider {
	switch t {
	case HTML:
		return &HTMLProvider{}
	default:
		return &DataProvider{}
	}
}
