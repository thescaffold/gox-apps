package auth

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors the auth-surface subset of TS app.controller.ts (there is
// no path prefix there — @Controller() at :101) — mounted under
// apps/identity by identity/app.module.go, so e.g. Login resolves to
// apps/identity/login.
var ROUTES = router.ForRoutes(
	router.Post("/initiate", []any{AuthController{}, "Initiate"}),
	router.Post("/verify", []any{AuthController{}, "Verify"}),
	router.Post("/secret", []any{AuthController{}, "Secret"}),
	router.Post("/login", []any{AuthController{}, "Login"}),
	router.Post("/switch", []any{AuthController{}, "SwitchWorkspace"}, requireUser()...),
	router.Post("/reset-secret/initiate", []any{AuthController{}, "ResetSecretInitiate"}),
	router.Post("/reset-secret/verify", []any{AuthController{}, "ResetSecretVerify"}),
	router.Post("/reset-secret/update", []any{AuthController{}, "ResetSecretUpdate"}),
	// PLAN M1-40: the browser shell's device surface.
	router.Get("/sessions", []any{AuthController{}, "Sessions"}),
	router.Post("/logout", []any{AuthController{}, "Logout"}, optionalUser()...),
	router.Get("/user-clients", []any{AuthController{}, "UserClients"}, requireUser()...),
	router.Post("/client-log/register", []any{AuthController{}, "RegisterClientLog"}, requireUser()...),
	router.Get("/attributes", []any{AuthController{}, "Attributes"}, requireUser()...),
	router.Patch("/attributes/:type", []any{AuthController{}, "UpdateAttribute"}, requireUser()...),
)
