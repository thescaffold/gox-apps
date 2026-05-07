package validator

import (
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

type JSONProvider struct{}

// Validate checks payload.Input against the JSON Schema in payload.Meta["schema"].
// Returns true when valid or when no schema is configured.
func (p *JSONProvider) Validate(payload *Payload) (bool, error) {
	raw, ok := payload.Meta["schema"]
	if !ok || raw == nil {
		return true, nil
	}

	schemaBytes, err := json.Marshal(raw)
	if err != nil {
		return false, fmt.Errorf("validator: marshal schema: %w", err)
	}

	inputBytes, err := json.Marshal(payload.Input)
	if err != nil {
		return false, fmt.Errorf("validator: marshal input: %w", err)
	}

	schemaLoader := gojsonschema.NewBytesLoader(schemaBytes)
	docLoader := gojsonschema.NewBytesLoader(inputBytes)

	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return false, fmt.Errorf("validator: json schema: %w", err)
	}
	return result.Valid(), nil
}
