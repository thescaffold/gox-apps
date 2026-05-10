package app

import ntxctx "github.com/thescaffold/gox-packages/libs/core/context"

type HealthDto struct{}

type RecentDto struct {
	NTX     ntxctx.NTXContext `context:"ntx"`
	Page    int               `form:"page"`
	PerPage int               `form:"perPage"`
}

type SearchDto struct {
	NTX     ntxctx.NTXContext `context:"ntx"`
	Query   string            `form:"query"   binding:"required"`
	Type    string            `form:"type"`
	Page    int               `form:"page"`
	PerPage int               `form:"perPage"`
}

type LookupDto struct {
	Group      string `form:"group"      binding:"required"`
	Service    string `form:"service"    binding:"required"`
	EntityName string `form:"entityName" binding:"required"`
	EntityId   string `form:"entityId"   binding:"required"`
}
