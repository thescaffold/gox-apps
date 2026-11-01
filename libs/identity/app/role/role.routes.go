package role

import (
	"github.com/awesome-goose/goose/modules/router"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/rbac"
)

// ROUTES mirrors TS @Controller('role') — singular path segment.
// /role/attach|detach|sync precede the Resource catch-all so they aren't
// shadowed by /:id.
var ROUTES = router.ForRoutes(
	router.Post("/role/attach", []any{RoleController{}, "Attach"}, rbac.Middleware()...),
	router.Post("/role/detach", []any{RoleController{}, "Detach"}, rbac.Middleware()...),
	router.Post("/role/sync", []any{RoleController{}, "Sync"}, rbac.Middleware()...),
	rbac.Secure(router.Resource("role", RoleController{}).All()),
)
