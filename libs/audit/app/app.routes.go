package app

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors ntx-apps/libs/audit/src/app.controller.ts: GET / is the hello
// endpoint, GET /activities returns a paginated list of described log rows.
// (LogModule contributes its own /log CRUD under the same apps/audit prefix.)
//
// Guards (PLAN U26): activities lists the caller's own workspace's audit trail, so it
// takes the workspace from the verified session (coreauth.Caller). Mounted with no
// middleware it always saw an empty workspace and listed nothing, signed in or not.
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Get("/activities", []any{AppController{}, "Activities"}, coreauth.Caller()...),
)
