package permission

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('permission') — singular path segment.
// /permission/attach|detach|sync precede the Resource catch-all so they
// aren't shadowed by /:id.
var ROUTES = router.ForRoutes(
	router.Post("/permission/attach", []any{PermissionController{}, "Attach"}),
	router.Post("/permission/detach", []any{PermissionController{}, "Detach"}),
	router.Post("/permission/sync", []any{PermissionController{}, "Sync"}),
	router.Resource("permission", PermissionController{}).All(),
)
