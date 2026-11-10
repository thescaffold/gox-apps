package message

import (
	"github.com/awesome-goose/goose/modules/router"
	"github.com/awesome-goose/goose/types"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('message') — singular path segment.
//
// Every route needs a verified caller, whose context the crud scoping reads (U26). Unless the
// host called ScopeToOwner, the resource is unscoped (every person's messages), so it is a
// platform admin's only; with ScopeToOwner each caller reaches their own messages.
var ROUTES = router.ForRoutes(
	coreauth.Guard(router.Resource("message", MessageController{}).All(), append(coreauth.Caller(), &scopedOrAdmin{})),
)

// scopedOrAdmin lets a caller through when the resource is scoped to its owner, and otherwise
// only a platform admin. It reads ScopeToOwner on every request, so it holds whatever order
// the host set it in.
type scopedOrAdmin struct{}

var _ types.Middleware = (*scopedOrAdmin)(nil)

func (g *scopedOrAdmin) Handle(ctx types.Context) error {
	if ownerScoped.Load() {
		return nil
	}
	return (&coreauth.AdminGuard{}).Handle(ctx)
}
