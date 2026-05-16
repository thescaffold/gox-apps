package app

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// HelloDto carries the request context for GET / (getHello). It has no body —
// the request context is the only input.
type HelloDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// ActivitiesDto is the query payload for GET /activities. Page/perPage are
// numeric query params; TS reads them as numbers (no validation), so empty
// or invalid values fall back to defaults applied in the handler.
type ActivitiesDto struct {
	Page    int               `query:"page"`
	PerPage int               `query:"perPage"`
	Ctx     ntxctx.NTXContext `context:"ntx"`
}
