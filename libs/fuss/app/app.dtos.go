package app

import ntxctx "github.com/thescaffold/gox-packages/libs/core/context"

// HelloDto carries the request context for GET / (getHello).
type HelloDto struct {
	NTX ntxctx.NTXContext `context:"ntx"`
}

// RecentDto carries the query params for GET /recent. TS reads them with
// @Query('page')/@Query('perPage') — no validation, defaults applied server-side.
type RecentDto struct {
	NTX     ntxctx.NTXContext `context:"ntx"`
	Page    int               `query:"page"`
	PerPage int               `query:"perPage"`
}

// SearchDto mirrors TS @Get('search') parameters. `query` is required; `type`
// defaults to "global" when absent.
type SearchDto struct {
	NTX     ntxctx.NTXContext `context:"ntx"`
	Query   string            `query:"query"   binding:"required"`
	Type    string            `query:"type"`
	Page    int               `query:"page"`
	PerPage int               `query:"perPage"`
}

// LookupDto mirrors TS @Get('lookup') parameters.
type LookupDto struct {
	NTX        ntxctx.NTXContext `context:"ntx"`
	Group      string            `query:"group"      binding:"required"`
	Service    string            `query:"service"    binding:"required"`
	EntityName string            `query:"entityName" binding:"required"`
	EntityId   string            `query:"entityId"   binding:"required"`
}
