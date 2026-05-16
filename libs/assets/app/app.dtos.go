package app

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// HelloDto carries the request context for GET / (getHello). It has no body —
// the request context is the only input.
type HelloDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// DynamicFileDto mirrors the TS @Get('dynamic') query params — every field is
// optional. The request context is included for translation/preference.
type DynamicFileDto struct {
	Id      *string           `query:"id"`
	Variant *string           `query:"variant"`
	Name    *string           `query:"name"`
	Colors  *string           `query:"colors"`
	Square  bool              `query:"square"`
	Size    int               `query:"size"`
	Raw     bool              `query:"raw"`
	Store   bool              `query:"store"`
	Ctx     ntxctx.NTXContext `context:"ntx"`
}
