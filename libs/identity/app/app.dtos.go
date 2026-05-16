package app

import ntxctx "github.com/thescaffold/gox-packages/libs/core/context"

// HelloDto carries request context for GET / (getHello).
type HelloDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}
