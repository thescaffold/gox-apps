package provider

import (
	"github.com/awesome-goose/goose/modules/router"
	"github.com/awesome-goose/goose/types"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('provider') — singular path segment.
//
// A provider row is a person's saved payment method (a card's authorisation token), linked by a
// payment verification. The owner may list, read, make primary and remove their own (the
// controller scopes every one of those to accounts in the caller's workspace and lets a caller
// change only `primary`); creating one by hand needs a platform admin (PLAN U26).
var ROUTES = router.ForRoutes(
	guard(router.Resource("provider", ProviderController{}).All()),
)

func guard(route types.Route) types.Route {
	children := make(types.Routes, 0, len(route.Children))
	for _, child := range route.Children {
		mws := coreauth.Caller()
		if tuple, ok := child.Handler.([]any); ok && len(tuple) == 2 && tuple[1] == "Create" {
			mws = coreauth.Admin()
		}
		child.Middlewares = append(mws, child.Middlewares...)
		children = append(children, child)
	}
	route.Children = children
	return route
}
