package apikey

import ntxctx "github.com/thescaffold/gox-packages/libs/core/context"

type IssueReq struct {
	Name        string            `json:"name" binding:"required"`
	Kind        string            `json:"kind"`
	Environment string            `json:"environment"`
	Scopes      []string          `json:"scopes"`
	ExpiresIn   int               `json:"expires_in_seconds"`
	Ctx         ntxctx.NTXContext `context:"ntx"`
}

type ListReq struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

type IDReq struct {
	Id    string            `param:"id"`
	Grace string            `query:"grace_seconds"`
	Ctx   ntxctx.NTXContext `context:"ntx"`
}
