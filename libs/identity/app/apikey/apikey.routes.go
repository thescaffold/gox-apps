package apikey

import (
	"github.com/awesome-goose/goose/modules/router"
	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
)

// Every route needs a verified user token: the tenant comes from its claims, never from headers.
// POST   /keys              issue a key (the secret is returned once)
// GET    /keys              list the workspace's keys (never their values)
// DELETE /keys/:id          revoke
// POST   /keys/:id/rotate   issue a replacement and retire the old key
var ROUTES = router.ForRoutes(
	router.Post("/keys", []any{ApiKeyController{}, "Issue"}, auth.RequireUser()...),
	router.Get("/keys", []any{ApiKeyController{}, "List"}, auth.RequireUser()...),
	router.Delete("/keys/:id", []any{ApiKeyController{}, "Revoke"}, auth.RequireUser()...),
	router.Post("/keys/:id/rotate", []any{ApiKeyController{}, "Rotate"}, auth.RequireUser()...),
)
