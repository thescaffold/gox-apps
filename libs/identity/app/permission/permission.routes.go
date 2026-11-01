package permission

import (
	"github.com/awesome-goose/goose/modules/router"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/rbac"
)

// ROUTES mirrors TS @Controller('permission') — singular path segment.
// /permission/attach|detach|sync precede the Resource catch-all so they
// aren't shadowed by /:id.
var ROUTES = router.ForRoutes(
	router.Post("/permission/attach", []any{PermissionController{}, "Attach"}, rbac.Middleware()...),
	router.Post("/permission/detach", []any{PermissionController{}, "Detach"}, rbac.Middleware()...),
	router.Post("/permission/sync", []any{PermissionController{}, "Sync"}, rbac.Middleware()...),
	rbac.Secure(router.Resource("permission", PermissionController{}).All()),
)
