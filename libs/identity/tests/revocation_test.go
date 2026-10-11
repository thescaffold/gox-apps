package tests

import (
	"context"
	"testing"
	"time"

	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/revocation"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

func TestRevocation_TokensIssuedBeforeAreRefusedAndLaterOnesAreNot(t *testing.T) {
	r := newRig(t)
	revocation.SetStore(&revocation.Store{DB: r.db})
	t.Cleanup(func() { revocation.SetStore(nil) })
	u, _ := r.person("rita")

	old := map[string]any{"sub": u.Id, "iatMs": float64(time.Now().Add(-time.Minute).UnixMilli())}
	legacy := map[string]any{"sub": u.Id} // issued before the claim existed
	if revocation.Check(old) || revocation.Check(legacy) {
		t.Fatal("nothing is revoked yet")
	}
	if ok, err := revocation.RevokeUser(context.Background(), u.Id, "test"); err != nil || !ok {
		t.Fatalf("revoke: %v %v", ok, err)
	}
	if !revocation.Check(old) || !revocation.Check(legacy) {
		t.Fatal("tokens from before the revocation must be refused")
	}
	fresh := map[string]any{"sub": u.Id, "iatMs": float64(time.Now().Add(time.Second).UnixMilli())}
	if revocation.Check(fresh) {
		t.Fatal("a token issued after the revocation must work")
	}
	// Someone else is unaffected, and a token with no subject is not our business.
	other, _ := r.person("sam")
	if revocation.Check(map[string]any{"sub": other.Id}) || revocation.Check(map[string]any{}) {
		t.Fatal("revocation must be per person")
	}
}

func TestRevocation_ClosingAnAccountEndsItEverywhere(t *testing.T) {
	r := newRig(t)
	revocation.SetStore(&revocation.Store{DB: r.db})
	t.Cleanup(func() { revocation.SetStore(nil) })
	u, _ := r.person("cleo")
	ctx := ntxctx.NTXContext{UserID: u.Id, ClientID: r.client}

	if rep := decode(t, r.ctl.Login(&auth.LoginDto{Ref: u.Email, Secret: testPassword, ClientId: r.client})); rep.code != 200 {
		t.Fatalf("before closing: %d", rep.code)
	}
	if rep := decode(t, r.ctl.CloseAccount(&auth.CloseAccountDto{Ctx: ctx, Confirm: "wrong@example.test"})); rep.code != 400 {
		t.Fatalf("a wrong confirmation: %d", rep.code)
	}
	if rep := decode(t, r.ctl.CloseAccount(&auth.CloseAccountDto{Ctx: ctxFor("nobody"), Confirm: u.Email})); rep.code != 401 {
		t.Fatalf("someone else's account: %d", rep.code)
	}
	if rep := decode(t, r.ctl.CloseAccount(&auth.CloseAccountDto{Ctx: ctx, Confirm: " " + u.Email + " "})); rep.code != 200 {
		t.Fatalf("close: %d", rep.code)
	}
	if rep := decode(t, r.ctl.Login(&auth.LoginDto{Ref: u.Email, Secret: testPassword, ClientId: r.client})); rep.code == 200 {
		t.Fatal("a closed account must not sign in")
	}
	if !revocation.Check(map[string]any{"sub": u.Id, "iatMs": float64(time.Now().Add(-time.Second).UnixMilli())}) {
		t.Fatal("tokens already issued must stop working")
	}
}

func ctxFor(id string) ntxctx.NTXContext { return ntxctx.NTXContext{UserID: id} }
