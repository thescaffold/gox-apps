package validator

type JSONProvider struct{}

func (p *JSONProvider) Validate(payload *Payload) (bool, error) {
	// stub: JSON schema validation not yet implemented
	return true, nil
}
