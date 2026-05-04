package mapper

type DataProvider struct{}

func (p *DataProvider) Map(payload *Payload) ([]any, error) {
	data, ok := payload.Input["data"]
	if !ok || data == nil {
		return nil, nil
	}
	switch v := data.(type) {
	case []any:
		return v, nil
	default:
		return []any{v}, nil
	}
}
