package role

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('role') — singular path segment.
// /role/attach|detach|sync precede the Resource catch-all so they aren't
// shadowed by /:id.
var ROUTES = router.ForRoutes(
	router.Post("/role/attach", []any{RoleController{}, "Attach"}),
	router.Post("/role/detach", []any{RoleController{}, "Detach"}),
	router.Post("/role/sync", []any{RoleController{}, "Sync"}),
	router.Resource("role", RoleController{}).All(),
)
