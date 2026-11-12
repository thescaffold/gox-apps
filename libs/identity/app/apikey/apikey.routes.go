package apikey

import (
	"github.com/awesome-goose/goose/modules/router"
	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
)

// Every route needs a verified user token: the tenant comes from its claims, never from headers.
// POST   /keys              issue a key (the secret is returned once)
// GET    /keys              list the caller's keys; owners and admins see the workspace's (never values)
// DELETE /keys/:id          revoke one's own key; owners and admins revoke anyone's
// POST   /keys/:id/rotate   issue a replacement for one's own key and retire the old one
// GET    /keys/policy       the workspace's maximum key life (V0-14)
// PUT    /keys/policy       owners and admins set it {maxDays}
var ROUTES = router.ForRoutes(
	router.Get("/keys/policy", []any{ApiKeyController{}, "Policy"}, auth.RequireUser()...),
	router.Put("/keys/policy", []any{ApiKeyController{}, "SetPolicy"}, auth.RequireUser()...),
	router.Post("/keys", []any{ApiKeyController{}, "Issue"}, auth.RequireUser()...),
	router.Get("/keys", []any{ApiKeyController{}, "List"}, auth.RequireUser()...),
	router.Delete("/keys/:id", []any{ApiKeyController{}, "Revoke"}, auth.RequireUser()...),
	router.Post("/keys/:id/rotate", []any{ApiKeyController{}, "Rotate"}, auth.RequireUser()...),
)
