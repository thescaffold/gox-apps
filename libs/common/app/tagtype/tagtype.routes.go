package tagtype

import (
	"github.com/awesome-goose/goose/modules/router"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// ROUTES mirrors TS @Controller('tag-type') — singular kebab-case path segment.
// POST /sync /attach /detach mirror the TS bulk-link endpoints (Phase 7.1).
var ROUTES = router.ForRoutes(
	router.Post("/tag-type/sync", []any{TagTypeController{}, "Sync"}, coreauth.Caller()...),
	router.Post("/tag-type/attach", []any{TagTypeController{}, "Attach"}, coreauth.Caller()...),
	router.Post("/tag-type/detach", []any{TagTypeController{}, "Detach"}, coreauth.Caller()...),
	coreauth.SecureCaller(router.Resource("tag-type", TagTypeController{}).All()),
)
