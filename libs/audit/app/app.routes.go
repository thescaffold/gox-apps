package app

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors ntx-apps/libs/audit/src/app.controller.ts: GET / is the hello
// endpoint, GET /activities returns a paginated list of described log rows.
// (LogModule contributes its own /log CRUD under the same apps/audit prefix.)
var ROUTES = router.ForRoutes(
	router.Get("/", []any{AppController{}, "GetHello"}),
	router.Get("/activities", []any{AppController{}, "Activities"}),
)
