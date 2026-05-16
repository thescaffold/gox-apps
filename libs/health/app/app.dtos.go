package app

import (
	"encoding/json"

	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// HelloDto carries the request context for GET / (getHello).
type HelloDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// PingDto mirrors TS app.dto.ts PingDto used by POST /ping.
type PingDto struct {
	Name  string            `json:"name"  binding:"required"`
	State string            `json:"state" binding:"required"`
	Meta  json.RawMessage   `json:"meta,omitempty"`
	Ctx   ntxctx.NTXContext `context:"ntx"`
}

// RegisterServiceDto mirrors TS CreateServiceDto used by POST /register.
type RegisterServiceDto struct {
	Name   string            `json:"name"             binding:"required"`
	Desc   *string           `json:"desc,omitempty"`
	State  *string           `json:"state,omitempty"`
	Type   *string           `json:"type,omitempty"`
	Status *string           `json:"status,omitempty"`
	Ctx    ntxctx.NTXContext `context:"ntx"`
}
