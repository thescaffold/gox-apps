package tagtype

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors TS @Controller('tag-type') — singular kebab-case path segment.
// POST /sync /attach /detach mirror the TS bulk-link endpoints (Phase 7.1).
var ROUTES = router.ForRoutes(
	router.Post("/tag-type/sync", []any{TagTypeController{}, "Sync"}),
	router.Post("/tag-type/attach", []any{TagTypeController{}, "Attach"}),
	router.Post("/tag-type/detach", []any{TagTypeController{}, "Detach"}),
	router.Resource("tag-type", TagTypeController{}).All(),
)
