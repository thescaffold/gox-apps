package apikey

import "github.com/awesome-goose/goose/modules/router"

// POST   /keys              issue a key (the secret is returned once)
// GET    /keys              list the workspace's keys (never their values)
// DELETE /keys/:id          revoke
// POST   /keys/:id/rotate   issue a replacement and retire the old key
var ROUTES = router.ForRoutes(
	router.Post("/keys", []any{ApiKeyController{}, "Issue"}),
	router.Get("/keys", []any{ApiKeyController{}, "List"}),
	router.Delete("/keys/:id", []any{ApiKeyController{}, "Revoke"}),
	router.Post("/keys/:id/rotate", []any{ApiKeyController{}, "Rotate"}),
)
