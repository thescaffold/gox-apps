package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/revocation"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
)

// POST /refresh is how a native app stays signed in. Every way a session can
// end must stop it, and every refusal must look the same.

func (r *rig) refresh(token string) reply {
	r.t.Helper()
	return decode(r.t, r.ctl.Refresh(&auth.RefreshDto{Token: token}))
}

func refreshTokenOf(t *testing.T, rep reply) string {
	t.Helper()
	tok, _ := rep.data()["refreshToken"].(string)
	if rep.code != 200 || tok == "" {
		t.Fatalf("sign-in gave no refresh token: %d %v", rep.code, rep.body)
	}
	return tok
}

func TestRefresh_ExchangesARefreshTokenForAnAccessTokenInTheSameWorkspace(t *testing.T) {
	r := newRig(t)
	ada, ws := r.person("ada")
	refresh := refreshTokenOf(t, r.login(ada, ""))

	rep := r.refresh(refresh)
	if rep.code != 200 {
		t.Fatalf("refresh: %d %v", rep.code, rep.body)
	}
	access, _ := rep.data()["accessToken"].(string)
	claims, err := coreauth.Verify(access, "changeme")
	if err != nil || claims["sub"] != ada.Id || claims["workspaceId"] != ws || claims["clientId"] != r.client {
		t.Fatalf("the new access token is %v (%v), want ada in her workspace on this client", claims, err)
	}
	if _, leaked := rep.data()["refreshToken"]; leaked {
		t.Fatal("refresh must not hand out another refresh token")
	}
	var n int64
	r.db.Table("IdentityTokens").Where(`"user_id" = ? AND "type" = 'refresh'`, ada.Id).Count(&n)
	if n != 1 {
		t.Fatalf("%d refresh records, want the 1 from sign-in (refreshing writes nothing)", n)
	}
}

func TestRefresh_RefusesWhatIsNotARefreshTokenOnRecord(t *testing.T) {
	r := newRig(t)
	ada, _ := r.person("ada")
	in := r.login(ada, "")
	access, _ := in.data()["accessToken"].(string)

	for name, tok := range map[string]string{
		"an access token": access, "garbage": "not.a.jwt", "an empty token": "",
	} {
		if rep := r.refresh(tok); rep.code != 401 {
			t.Errorf("%s: %d, want 401", name, rep.code)
		}
	}
}

func TestRefresh_EndsWithSignOutSuspensionAndSignOutEverywhere(t *testing.T) {
	r := newRig(t)
	revocation.SetStore(&revocation.Store{DB: r.db})
	t.Cleanup(func() { revocation.SetStore(nil) })

	// signed out: the refresh record is deleted
	out, _ := r.person("out")
	outRefresh := refreshTokenOf(t, r.login(out, ""))
	r.db.Exec(`DELETE FROM "IdentityTokens" WHERE "user_id" = ?`, out.Id)
	if rep := r.refresh(outRefresh); rep.code != 401 {
		t.Errorf("after sign-out: %d, want 401", rep.code)
	}

	// suspended
	sus, _ := r.person("sus")
	susRefresh := refreshTokenOf(t, r.login(sus, ""))
	if rep := r.refresh(susRefresh); rep.code != 200 {
		t.Fatalf("before suspension: %d", rep.code)
	}
	r.db.Exec(`INSERT INTO "IdentityAttributes" (id, created_at, updated_at, user_id, key, value) VALUES (?, now(), now(), ?, 'suspended_at', 'x')`, uuid.NewString(), sus.Id)
	if rep := r.refresh(susRefresh); rep.code != 401 {
		t.Errorf("suspended: %d, want 401", rep.code)
	}

	// signed out everywhere
	all, _ := r.person("all")
	allRefresh := refreshTokenOf(t, r.login(all, ""))
	if _, err := revocation.RevokeUser(context.Background(), all.Id, "test"); err != nil {
		t.Fatal(err)
	}
	if rep := r.refresh(allRefresh); rep.code != 401 {
		t.Errorf("after sign-out everywhere: %d, want 401", rep.code)
	}
}

func TestRefresh_RefusesAWorkspaceThePersonHasLeft(t *testing.T) {
	r := newRig(t)
	ada, ws := r.person("ada")
	refresh := refreshTokenOf(t, r.login(ada, ""))
	r.db.Exec(`DELETE FROM "IdentityUserClientWorkspaces" WHERE "user_id" = ? AND "workspace_id" = ?`, ada.Id, ws)
	if rep := r.refresh(refresh); rep.code != 401 {
		t.Fatalf("left the workspace: %d, want 401", rep.code)
	}
}
