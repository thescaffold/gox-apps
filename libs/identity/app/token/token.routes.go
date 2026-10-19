package token

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('token') — singular path segment.
// POST /token/user precedes the CRUD Resource so it isn't shadowed by /:id.
var ROUTES = router.ForRoutes(
	router.Post("/token/user", []any{TokenController{}, "User"}),
	coreauth.Secure(router.Resource("token", TokenController{}).All()),
)
