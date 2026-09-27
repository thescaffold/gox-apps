package auth

import (
	"os"

	"github.com/awesome-goose/goose/types"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// jwtSecret mirrors identity/pkg's AuthService.jwtSecret() exactly (same env
// var, same "changeme" fallback) — duplicated rather than exported from pkg
// because pkg's copy is a private method on AuthService, and tokens signed
// by AuthService.Login/IssueTokensByHandle must be verified with the same
// secret this middleware checks them against.
func jwtSecret() string {
	if v := os.Getenv("JWT_SECRET"); v != "" {
		return v
	}
	return "changeme"
}

// claimsToContext runs after coreauth.AuthMiddleware has verified the Bearer
// JWT and stored its claims; it derives a trusted NTXContext from those
// verified claims and stores it the way the `context:"ntx"` DTO tag expects
// (ntxctx.Get/Set, gox-packages/core/context). This is the safe half of
// wiring authenticated identity into a handler: gox-packages/core/context's
// own Middleware parses x-ntx-* headers directly off the incoming request
// with no verification at all — fine for trusted service-to-service calls,
// but never safe to mount on a public-facing route, since any client could
// set x-ntx-user-id themselves and impersonate another user. Deriving
// NTXContext.UserID/ClientID from AuthMiddleware's already-verified JWT
// claims instead removes that hole for the one route pair (GET/PATCH
// /attributes) that needs a real authenticated user.
type claimsToContext struct{}

func (m *claimsToContext) Handle(ctx types.Context) error {
	claims := coreauth.GetClaims(ctx)
	ntx := ntxctx.NTXContext{}
	if claims != nil {
		if v, ok := claims["sub"].(string); ok {
			ntx.UserID = v
		}
		if v, ok := claims["clientId"].(string); ok {
			ntx.ClientID = v
		}
	}
	ntxctx.Set(ctx, ntx)
	return nil
}

// requireUser is the middleware pair every authenticated route in this
// package applies: verify the Bearer JWT, then translate its claims into
// the NTXContext handlers read via dto.Ctx.
func requireUser() []types.Middleware {
	return []types.Middleware{
		&coreauth.AuthMiddleware{Secret: jwtSecret()},
		&claimsToContext{},
	}
}
