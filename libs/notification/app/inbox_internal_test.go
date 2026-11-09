package app

import (
	"testing"

	"github.com/awesome-goose/goose/io/input"
	goosetest "github.com/awesome-goose/goose/testing"
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

func TestInboxReadDtoBindsThePathID(t *testing.T) {
	ctx := goosetest.NewMockContext()
	ctx.MockRequest().WithParams(map[string]string{"id": "m1"})
	var dto InboxReadDto
	if err := input.NewInput(ctx).Populate(&dto); err != nil {
		t.Fatal(err)
	}
	if dto.ID != "m1" {
		t.Fatalf("id = %q", dto.ID)
	}
}
