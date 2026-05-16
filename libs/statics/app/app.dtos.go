package app

import ntxctx "github.com/thescaffold/gox-packages/libs/core/context"

// HelloDto carries request context for GET / (getHello).
type HelloDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// FilterByKeyDto mirrors @Get('filter/:key').
type FilterByKeyDto struct {
	Key      string            `path:"key"`
	ParentId *string           `query:"parentId"`
	Ctx      ntxctx.NTXContext `context:"ntx"`
}

// FindByCodeDto mirrors @Get('code/:code').
type FindByCodeDto struct {
	Code     string            `path:"code"`
	ParentId *string           `query:"parentId"`
	Ctx      ntxctx.NTXContext `context:"ntx"`
}

// FindByValueDto mirrors @Get('value/:value').
type FindByValueDto struct {
	Value    string            `path:"value"`
	ParentId *string           `query:"parentId"`
	Ctx      ntxctx.NTXContext `context:"ntx"`
}
