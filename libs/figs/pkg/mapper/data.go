package mapper

type DataProvider struct{}

// Map mirrors TS DataService.map which returns `[headers, data]`: headers come
// from input.templateObject.headers (default empty), data is input.data. The
// converter then writes the header row followed by the data rows.
func (p *DataProvider) Map(payload *Payload) ([]any, error) {
	var headers any = []any{}
	if to, ok := payload.Input["templateObject"].(map[string]any); ok {
		if h, ok := to["headers"]; ok && h != nil {
			headers = h
		}
	}
	return []any{headers, payload.Input["data"]}, nil
}
