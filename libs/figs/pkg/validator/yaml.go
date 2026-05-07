package validator

import (
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

type YAMLProvider struct{}

// Validate checks payload.Input against the YAML schema in payload.Meta["schema"].
// The schema is treated as a JSON Schema document (YAML syntax accepted).
// Returns true when valid or when no schema is configured.
func (p *YAMLProvider) Validate(payload *Payload) (bool, error) {
	raw, ok := payload.Meta["schema"]
	if !ok || raw == nil {
		return true, nil
	}

	// Re-encode the schema map to YAML bytes, then unmarshal to normalised any.
	schemaYAML, err := yaml.Marshal(raw)
	if err != nil {
		return false, fmt.Errorf("validator: marshal yaml schema: %w", err)
	}
	var schemaDoc any
	if err := yaml.Unmarshal(schemaYAML, &schemaDoc); err != nil {
		return false, fmt.Errorf("validator: unmarshal yaml schema: %w", err)
	}
	schemaBytes, err := json.Marshal(normalise(schemaDoc))
	if err != nil {
		return false, fmt.Errorf("validator: convert schema to json: %w", err)
	}

	inputBytes, err := json.Marshal(normalise(payload.Input))
	if err != nil {
		return false, fmt.Errorf("validator: convert input to json: %w", err)
	}

	result, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schemaBytes),
		gojsonschema.NewBytesLoader(inputBytes),
	)
	if err != nil {
		return false, fmt.Errorf("validator: yaml schema: %w", err)
	}
	return result.Valid(), nil
}

// normalise converts map[any]any (produced by gopkg.in/yaml.v3) to map[string]any
// so it can be JSON-marshalled.
func normalise(v any) any {
	switch val := v.(type) {
	case map[any]any:
		out := make(map[string]any, len(val))
		for k, vv := range val {
			out[fmt.Sprintf("%v", k)] = normalise(vv)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, vv := range val {
			out[k] = normalise(vv)
		}
		return out
	case []any:
		for i, item := range val {
			val[i] = normalise(item)
		}
		return val
	default:
		return v
	}
}
