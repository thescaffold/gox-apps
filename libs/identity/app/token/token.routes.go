package token

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('token') — singular path segment.
// POST /token/user precedes the CRUD Resource so it isn't shadowed by /:id.
var ROUTES = router.ForRoutes(
	router.Post("/token/user", []any{TokenController{}, "User"}),
	router.Resource("token", TokenController{}).All(),
)
