package validator

type YAMLProvider struct{}

func (p *YAMLProvider) Validate(payload *Payload) (bool, error) {
	// stub: YAML schema validation not yet implemented
	return true, nil
}
