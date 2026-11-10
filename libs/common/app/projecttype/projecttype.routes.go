package projecttype

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('project-type') — singular kebab-case segment.
// POST /sync /attach /detach mirror the TS bulk-link endpoints (Phase 7.1).
var ROUTES = router.ForRoutes(
	router.Post("/project-type/sync", []any{ProjectTypeController{}, "Sync"}, coreauth.Caller()...),
	router.Post("/project-type/attach", []any{ProjectTypeController{}, "Attach"}, coreauth.Caller()...),
	router.Post("/project-type/detach", []any{ProjectTypeController{}, "Detach"}, coreauth.Caller()...),
	coreauth.SecureCaller(router.Resource("project-type", ProjectTypeController{}).All()),
)
