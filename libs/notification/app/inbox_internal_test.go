package app

import (
	"testing"

	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// The inbox routes take the caller only from a verified session.
func TestInboxGuardIsSessionThenClaims(t *testing.T) {
	g := inboxGuard()
	if len(g) != 2 {
		t.Fatalf("%d middlewares", len(g))
	}
	if _, ok := g[0].(*coreauth.AuthMiddleware); !ok {
		t.Fatalf("first middleware is %T, want the session check", g[0])
	}
	if _, ok := g[1].(*coreauth.FromClaims); !ok {
		t.Fatalf("second middleware is %T, want FromClaims", g[1])
	}
}
