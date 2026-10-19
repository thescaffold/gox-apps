package tests

import (
	"testing"

	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// The browser sign-in hop: a client app with no session starts it at oauth/web/initiate, the sign-in
// screens mint a code for the signed-in person, and the client redeems it, once, for a session.
// Without these routes the sign-in page never finished loading (UP-9).

func TestWebInitiateAnswersWithASessionAStateAndAScope(t *testing.T) {
	r := newRig(t)
	rep := decode(t, r.ctl.OAuthWebInitiate(&auth.OAuthInitiateDto{ClientId: r.client}))
	if rep.code != 200 {
		t.Fatalf("HTTP %d %v", rep.code, rep.body)
	}
	d := rep.data()
	for _, k := range []string{"sessionId", "state", "scope", "id", "redirectUrl"} {
		if s, _ := d[k].(string); s == "" {
			t.Errorf("%s is empty in %v", k, d)
		}
	}
	if d["id"] != r.client {
		t.Errorf("id = %v, want the client id %s", d["id"], r.client)
	}
}

func TestWebInitiateKeepsTheStateAndScopeItIsGiven(t *testing.T) {
	r := newRig(t)
	d := decode(t, r.ctl.OAuthServerInitiate(&auth.OAuthInitiateDto{ClientId: r.client, State: "abc", Scope: "mail"})).data()
	if d["state"] != "abc" || d["scope"] != "mail" {
		t.Fatalf("got %v", d)
	}
}

func TestACodeIsRedeemedOnceForTheSignedInPersonsSession(t *testing.T) {
	r := newRig(t)
	u, ws := r.person("ada")
	sess := decode(t, r.ctl.OAuthWebInitiate(&auth.OAuthInitiateDto{ClientId: r.client})).data()["sessionId"].(string)

	ctx := signedIn(u.Id, ws)
	code := decode(t, r.ctl.OAuthAuthCode(&auth.OAuthAuthCodeDto{Ctx: ctx, SessionId: sess})).data()["authCode"].(string)

	// The start page sends the code in `state` and the state in `authCode`.
	redeem := func() reply {
		return decode(t, r.ctl.OAuthAccessToken(&auth.OAuthAccessTokenDto{ClientId: r.client, SessionId: sess, State: code, AuthCode: "x"}))
	}
	first := redeem()
	if first.code != 200 {
		t.Fatalf("redeem: HTTP %d %v", first.code, first.body)
	}
	d := first.data()
	if s, _ := d["accessToken"].(string); s == "" {
		t.Errorf("no access token in %v", d)
	}
	if user, _ := d["user"].(map[string]any); user["id"] != u.Id {
		t.Errorf("signed in as %v, want %s", user, u.Id)
	}
	if again := redeem(); again.code == 200 {
		t.Errorf("a code must work once; the second redeem answered %d", again.code)
	}
}

func TestACodeCannotBeMintedWithoutASignedInPersonOrForAnotherSession(t *testing.T) {
	r := newRig(t)
	sess := decode(t, r.ctl.OAuthWebInitiate(&auth.OAuthInitiateDto{ClientId: r.client})).data()["sessionId"].(string)
	if rep := decode(t, r.ctl.OAuthAuthCode(&auth.OAuthAuthCodeDto{SessionId: sess})); rep.code != 401 {
		t.Errorf("no person: HTTP %d, want 401", rep.code)
	}
	u, ws := r.person("bob")
	if rep := decode(t, r.ctl.OAuthAuthCode(&auth.OAuthAuthCodeDto{Ctx: signedIn(u.Id, ws), SessionId: "no-such-session"})); rep.code != 400 {
		t.Errorf("unknown session: HTTP %d, want 400", rep.code)
	}
	if rep := decode(t, r.ctl.OAuthAccessToken(&auth.OAuthAccessTokenDto{SessionId: sess, AuthCode: "guess"})); rep.code == 200 {
		t.Errorf("a guessed code was accepted")
	}
}

func signedIn(userID, workspaceID string) ntxctx.NTXContext {
	return ntxctx.NTXContext{UserID: userID, WorkspaceID: workspaceID}
}

// The shell's USER_LOGIN effect takes a sign-in response whole: it needs the account's `id` and a
// token under `workspaces[].workspace.token`. Without them the browser signed in with no token.
func TestEverySignInAnswersInTheShapeTheShellReads(t *testing.T) {
	r := newRig(t)
	u, ws := r.person("eve")
	d := r.login(u, "").data()
	if d["id"] != u.Id {
		t.Errorf("id = %v, want %s", d["id"], u.Id)
	}
	if s, _ := d["token"].(string); s == "" {
		t.Errorf("no token for the first workspace in %v", d)
	}
	links, _ := d["workspaces"].([]any)
	if len(links) != 1 {
		t.Fatalf("workspaces = %v", d["workspaces"])
	}
	link := links[0].(map[string]any)
	if link["id"] != ws || link["workspaceId"] != ws || link["userId"] != u.Id {
		t.Errorf("link = %v", link)
	}
	if tok, _ := link["workspace"].(map[string]any)["token"].(string); tok == "" {
		t.Errorf("no workspace token in %v", link)
	}
	if s, _ := d["accessToken"].(string); s == "" {
		t.Errorf("the existing accessToken field is gone")
	}
}
