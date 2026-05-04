package converter

type ProviderType string

const (
	Excel ProviderType = "excel"
	CSV   ProviderType = "csv"
	PDF   ProviderType = "pdf"
)

type Response struct {
	Buffer    []byte
	Mime      string
	Encoding  string
	Extension string
}

type Provider interface {
	Convert(args ...any) (*Response, error)
}

type Service struct{}

func (s *Service) Use(t ProviderType) Provider {
	switch t {
	case Excel:
		return &ExcelProvider{}
	case PDF:
		return &PDFProvider{}
	default:
		return &CSVProvider{}
	}
}
