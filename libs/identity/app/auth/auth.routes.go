package auth

import "github.com/awesome-goose/goose/modules/router"

// ROUTES mirrors the auth-surface subset of TS app.controller.ts (there is
// no path prefix there — @Controller() at :101) — mounted under
// apps/identity by identity/app.module.go, so e.g. Login resolves to
// apps/identity/login.
var ROUTES = router.ForRoutes(
	router.Post("/oauth/web/initiate", []any{AuthController{}, "OAuthWebInitiate"}),
	router.Post("/oauth/server/initiate", []any{AuthController{}, "OAuthServerInitiate"}),
	router.Post("/oauth/device/initiate", []any{AuthController{}, "OAuthDeviceInitiate"}),
	router.Post("/oauth/server/auth-code", []any{AuthController{}, "OAuthAuthCode"}, requireUser()...),
	router.Post("/oauth/web/access-token", []any{AuthController{}, "OAuthAccessToken"}),
	router.Post("/oauth/server/access-token", []any{AuthController{}, "OAuthAccessToken"}),
	router.Post("/oauth/device/access-token", []any{AuthController{}, "OAuthAccessToken"}),
	router.Post("/initiate", []any{AuthController{}, "Initiate"}),
	router.Post("/verify", []any{AuthController{}, "Verify"}),
	router.Post("/secret", []any{AuthController{}, "Secret"}),
	router.Post("/login", []any{AuthController{}, "Login"}),
	// A native app's way to stay signed in: refresh token in, new access token out.
	router.Post("/refresh", []any{AuthController{}, "Refresh"}),
	// Passwordless sign-in: ask for a code, verify it, get tokens (creates the account on first use).
	router.Post("/otp", []any{AuthController{}, "OtpRequest"}),
	router.Post("/otp/verify", []any{AuthController{}, "OtpVerify"}),
	// Enterprise single sign-on (OpenID Connect). start/callback/exchange are
	// the browser flow and carry no session; connection is the admin's.
	router.Post("/sso/discover", []any{AuthController{}, "SsoDiscover"}),
	router.Post("/sso/start", []any{AuthController{}, "SsoStart"}),
	router.Get("/sso/callback", []any{AuthController{}, "SsoCallback"}),
	router.Post("/sso/exchange", []any{AuthController{}, "SsoExchange"}),
	router.Get("/sso/saml/:workspace", []any{AuthController{}, "SsoSamlMetadata"}),
	router.Post("/sso/saml/:workspace/acs", []any{AuthController{}, "SsoSamlAcs"}),
	router.Post("/sso/recovery/start", []any{AuthController{}, "SsoRecoveryStart"}),
	router.Post("/sso/recovery/verify", []any{AuthController{}, "SsoRecoveryVerify"}),
	router.Get("/sso/connection", []any{AuthController{}, "SsoGetConnection"}, requireUser()...),
	router.Put("/sso/connection", []any{AuthController{}, "SsoPutConnection"}, requireUser()...),
	router.Delete("/sso/connection", []any{AuthController{}, "SsoDeleteConnection"}, requireUser()...),
	router.Post("/secret/change", []any{AuthController{}, "ChangeSecret"}, requireUser()...),
	// Two-factor sign-in (U-S20, mfa.go): the second step, then the person's own setup and the policy.
	router.Post("/mfa/verify", []any{AuthController{}, "MfaVerify"}),
	router.Post("/mfa/verify/key", []any{AuthController{}, "MfaKeyOptions"}),
	router.Get("/mfa", []any{AuthController{}, "MfaStatus"}, requireUser()...),
	router.Post("/mfa/challenge", []any{AuthController{}, "MfaStepUp"}, requireUser()...),
	router.Post("/mfa/totp", []any{AuthController{}, "MfaTotpStart"}, requireUser()...),
	router.Post("/mfa/totp/confirm", []any{AuthController{}, "MfaTotpConfirm"}, requireUser()...),
	router.Post("/mfa/keys", []any{AuthController{}, "MfaKeyStart"}, requireUser()...),
	router.Post("/mfa/keys/confirm", []any{AuthController{}, "MfaKeyConfirm"}, requireUser()...),
	router.Post("/mfa/factors/:id/remove", []any{AuthController{}, "MfaRemove"}, requireUser()...),
	router.Post("/mfa/recovery", []any{AuthController{}, "MfaRecovery"}, requireUser()...),
	router.Get("/mfa/policy", []any{AuthController{}, "MfaPolicy"}, requireUser()...),
	router.Put("/mfa/policy", []any{AuthController{}, "MfaSetPolicy"}, requireUser()...),
	router.Get("/my/workspaces", []any{AuthController{}, "MyWorkspaces"}, requireUser()...),
	router.Post("/my/workspaces", []any{AuthController{}, "CreateMyWorkspace"}, requireUser()...),
	router.Patch("/my/workspaces/:id", []any{AuthController{}, "UpdateMyWorkspace"}, requireUser()...),
	router.Post("/account/close", []any{AuthController{}, "CloseAccount"}, requireUser()...),
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
