package app

import ntxctx "github.com/thescaffold/gox-packages/libs/core/context"

// HelloDto carries request context for GET / (getHello).
type HelloDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// GetCurrencyDto mirrors @Get('currency/:currency').
type GetCurrencyDto struct {
	Currency string            `uri:"currency"   binding:"required"`
	Ctx      ntxctx.NTXContext `context:"ntx"`
}

// GetLocationDto mirrors @Get('location') — TS reads ?ip= query, falling back
// to the request IP header. The Ctx field flows preference + identity through
// for i18n; HeaderIP captures the X-Forwarded-For / Remote IP fallback that TS
// resolves via @GetIP().
type GetLocationDto struct {
	Ip       string            `form:"ip"`
	HeaderIP string            `header:"x-forwarded-for"`
	Ctx      ntxctx.NTXContext `context:"ntx"`
}

// Sync/Attach/Detach DTOs (TagActionDto, ProjectActionDto in TS
// app.dto.ts) live in their respective sub-packages (tagtype, projecttype) —
// putting them here would create a circular import since each sub-package
// already imports the parent `app` package for nothing.
