package projecttype

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('project-type') — singular kebab-case segment.
// POST /sync /attach /detach mirror the TS bulk-link endpoints (Phase 7.1).
var ROUTES = router.ForRoutes(
	router.Post("/project-type/sync", []any{ProjectTypeController{}, "Sync"}),
	router.Post("/project-type/attach", []any{ProjectTypeController{}, "Attach"}),
	router.Post("/project-type/detach", []any{ProjectTypeController{}, "Detach"}),
	coreauth.Secure(router.Resource("project-type", ProjectTypeController{}).All()),
)
