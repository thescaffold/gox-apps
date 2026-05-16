package app

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// HelloDto carries the request context for GET / (getHello). It has no body —
// the request context is the only input.
type HelloDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// NowDto carries the request context for GET /now.
type NowDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// RegisterDto is the body of POST /register. Mirrors TS CreateRouteDto (the
// payload is then materialised once as a Http and once as a Ws route).
type RegisterDto struct {
	Group    string            `json:"group"           binding:"required"`
	Service  string            `json:"service"         binding:"required"`
	Type     string            `json:"type"`
	Name     string            `json:"name"            binding:"required"`
	Desc     *string           `json:"desc,omitempty"`
	Upstream string            `json:"upstream"        binding:"required"`
	Status   *string           `json:"status,omitempty"`
	Ctx      ntxctx.NTXContext `context:"ntx"`
}
