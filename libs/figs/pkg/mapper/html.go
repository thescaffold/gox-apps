package mapper

type HTMLProvider struct{}

func (p *HTMLProvider) Map(payload *Payload) ([]any, error) {
	templateBody, _ := payload.Input["templateBody"].(string)
	if templateBody == "" {
		return nil, nil
	}
	return []any{templateBody}, nil
}
