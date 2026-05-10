package app

type Payload struct {
	Meta   map[string]any `json:"meta"`
	Input  map[string]any `json:"input"`
	Output map[string]any `json:"output"`
}

type SaveFileDto struct {
	Name      string         `param:"name"`
	MetaRaw   map[string]any `json:"meta"`
	InputRaw  map[string]any `json:"input"`
	OutputRaw map[string]any `json:"output"`
}

type HealthDto struct{}
