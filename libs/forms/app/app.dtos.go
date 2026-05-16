package app

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// HelloDto carries the request context for GET / (getHello).
type HelloDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// GetFormDto carries the :id path param for GET /one/:id.
type GetFormDto struct {
	Id  string            `param:"id"  binding:"required"`
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// SaveFormDto carries :id + free-form key/value body for POST /one/:id.
// TS reads the request body as `Record<string, any>` directly — no DTO schema.
type SaveFormDto struct {
	Id   string            `param:"id"      binding:"required"`
	Body map[string]any    `json:",inline"`
	Ctx  ntxctx.NTXContext `context:"ntx"`
}
