package app

import ntxctx "github.com/thescaffold/gox-packages/libs/core/context"

// Payload mirrors TS ntx-apps/libs/figs/src/app.config.ts Payload — three
// loosely-typed nested maps. Validation runs inside the chosen
// validator/mapper/store providers, so we keep them as map[string]any here.
type Payload struct {
	Meta   map[string]any `json:"meta"`
	Input  map[string]any `json:"input"`
	Output map[string]any `json:"output"`
}

// HelloDto carries the request context for GET / (getHello).
type HelloDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// SaveFileDto mirrors TS @Post(':name') saveFile(@Param('name'),@Body PayloadDto).
type SaveFileDto struct {
	Name      string            `param:"name"`
	MetaRaw   map[string]any    `json:"meta"`
	InputRaw  map[string]any    `json:"input"`
	OutputRaw map[string]any    `json:"output"`
	Ctx       ntxctx.NTXContext `context:"ntx"`
}
