package converter

import "fmt"

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

// normalizeRows coerces a converter argument into table rows. It handles the
// JSON-decoded shapes the data mapper produces — []any of scalars (a single row,
// e.g. the headers) and []any of []any (multiple rows) — as well as the
// pre-typed []string / [][]string. The second return is false for non-tabular
// args (e.g. a rendered HTML string), so callers can fall back to their default.
func normalizeRows(arg any) ([][]string, bool) {
	switch v := arg.(type) {
	case []string:
		return [][]string{v}, true
	case [][]string:
		return v, true
	case []any:
		if len(v) > 0 {
			if _, isRow := v[0].([]any); isRow {
				rows := make([][]string, 0, len(v))
				for _, r := range v {
					rows = append(rows, cellsOf(r))
				}
				return rows, true
			}
		}
		return [][]string{cellsOf(v)}, true
	}
	return nil, false
}

// cellsOf stringifies a row ([]any of cells) into []string.
func cellsOf(r any) []string {
	arr, ok := r.([]any)
	if !ok {
		return []string{fmt.Sprintf("%v", r)}
	}
	out := make([]string, len(arr))
	for i, c := range arr {
		out[i] = fmt.Sprintf("%v", c)
	}
	return out
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
