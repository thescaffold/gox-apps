package tests

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	"github.com/thescaffold/gox-apps/libs/identity/app/user"
	identityucw "github.com/thescaffold/gox-apps/libs/identity/app/userclientworkspace"
	"github.com/thescaffold/gox-apps/libs/identity/pkg"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/mfa"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// clock is the test's time: every TOTP code is computed for it, and moving it
// on a step makes a fresh code.
type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }
func (c *clock) step()          { c.t = c.t.Add(30 * time.Second) }

type mfaRig struct {
	*rig
	clk *clock
	svc *mfa.Service
}

func newMfaRig(t *testing.T) *mfaRig {
	t.Helper()
	r := newRig(t)
	t.Setenv("IDENTITY_TOKEN_KEY", "0123456789abcdef0123456789abcdef")
	clk := &clock{t: time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)}
	svc := r.ctl.MfaForTest()
	svc.Now = clk.now
	return &mfaRig{rig: r, clk: clk, svc: svc}
}

func ctxOf(u *user.User, ws string) ntxctx.NTXContext {
	return ntxctx.NTXContext{UserID: u.Id, WorkspaceID: ws}
}

func (r *mfaRig) code(secret string) string {
	r.t.Helper()
	c, err := totp.GenerateCodeCustom(secret, r.clk.now(), totp.ValidateOpts{Period: 30, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
	if err != nil {
		r.t.Fatal(err)
	}
	return c
}

// enroll turns on an authenticator app for u and returns its secret and the recovery codes.
func (r *mfaRig) enroll(u *user.User, ws string) (string, []string) {
	r.t.Helper()
	start := decode(r.t, r.ctl.MfaTotpStart(&auth.MfaProofDto{Ctx: ctxOf(u, ws)}))
	if start.code != 200 {
		r.t.Fatalf("totp start: %d %v", start.code, start.body)
	}
	secret, _ := start.data()["secret"].(string)
	factor, _ := start.data()["factorId"].(string)
	if secret == "" || factor == "" {
		r.t.Fatalf("totp start: %v", start.data())
	}
	if qr, _ := start.data()["qr"].(string); len(qr) < 100 {
		r.t.Fatalf("no QR code: %q", qr)
	}
	confirm := decode(r.t, r.ctl.MfaTotpConfirm(&auth.MfaTotpConfirmDto{Ctx: ctxOf(u, ws), FactorId: factor, Code: r.code(secret)}))
	if confirm.code != 200 {
		r.t.Fatalf("totp confirm: %d %v", confirm.code, confirm.body)
	}
	raw, _ := confirm.data()["recoveryCodes"].([]any)
	codes := make([]string, 0, len(raw))
	for _, c := range raw {
		codes = append(codes, c.(string))
	}
	r.clk.step()
	return secret, codes
}

func (r *mfaRig) verify(challenge string, p auth.MfaVerifyDto) reply {
	r.t.Helper()
	p.Challenge = challenge
	return decode(r.t, r.ctl.MfaVerify(&p))
}

func (r *mfaRig) deviceSessions(u *user.User) int64 {
	var n int64
	r.db.Raw(`SELECT COUNT(*) FROM "IdentityDeviceSessions" WHERE "user_id" = ?`, u.Id).Scan(&n)
	return n
}

func challengeOf(t *testing.T, rep reply) string {
	t.Helper()
	d := rep.data()
	if rep.code != 200 || d["mfaRequired"] != true || d["accessToken"] != nil || d["token"] != nil {
		t.Fatalf("want a second step and no tokens: %d %v", rep.code, rep.body)
	}
	if _, ok := rep.headers["Set-Cookie"]; ok {
		t.Fatalf("no device cookie before the second factor: %v", rep.headers)
	}
	c, _ := d["challenge"].(string)
	if c == "" {
		t.Fatalf("no challenge: %v", d)
	}
	return c
}

func TestMfa_AuthenticatorAppGuardsPasswordSignIn(t *testing.T) {
	r := newMfaRig(t)
	ada, ws := r.person("ada")

	// Without 2FA nothing changes.
	if first := r.login(ada, ""); first.code != 200 || first.data()["accessToken"] == nil {
		t.Fatalf("plain sign-in: %d %v", first.code, first.body)
	}
	if mf, _ := r.login(ada, "").data()["mfa"].(map[string]any); mf["enabled"] != false {
		t.Fatalf("the session says whether 2FA is on: %v", mf)
	}

	// A wrong first code does not turn it on.
	start := decode(t, r.ctl.MfaTotpStart(&auth.MfaProofDto{Ctx: ctxOf(ada, ws)}))
	bad := decode(t, r.ctl.MfaTotpConfirm(&auth.MfaTotpConfirmDto{Ctx: ctxOf(ada, ws), FactorId: start.data()["factorId"].(string), Code: "000000"}))
	if bad.code != 401 || bad.body["code"] != "mfa_invalid" {
		t.Fatalf("wrong code confirms: %d %v", bad.code, bad.body)
	}
	if on, _ := r.svc.Enrolled(context.Background(), ada.Id); on {
		t.Fatal("enrolled after a wrong code")
	}

	secret, codes := r.enroll(ada, ws)
	if len(codes) != mfa.RecoveryCount {
		t.Fatalf("%d recovery codes", len(codes))
	}

	before := r.deviceSessions(ada)
	ch := challengeOf(t, r.login(ada, ""))
	if r.deviceSessions(ada) != before {
		t.Fatal("a device session was created before the second factor")
	}
	if wrong := r.verify(ch, auth.MfaVerifyDto{Code: "123456"}); wrong.code != 401 || wrong.body["code"] != "mfa_invalid" {
		t.Fatalf("wrong code: %d %v", wrong.code, wrong.body)
	}
	good := r.verify(ch, auth.MfaVerifyDto{Code: r.code(secret)})
	if good.code != 200 || good.data()["accessToken"] == nil || good.data()["mfaMethod"] != "totp" {
		t.Fatalf("right code: %d %v", good.code, good.body)
	}
	deviceFrom(t, good)
	if r.deviceSessions(ada) != before+1 {
		t.Fatal("no device session after the second factor")
	}
	if mf, _ := good.data()["mfa"].(map[string]any); mf["enabled"] != true {
		t.Fatalf("session payload: %v", mf)
	}

	// The same code (same time step) is never accepted twice, even on a new sign-in.
	replay := r.verify(challengeOf(t, r.login(ada, "")), auth.MfaVerifyDto{Code: r.code(secret)})
	if replay.code != 401 {
		t.Fatalf("a used code was accepted again: %d %v", replay.code, replay.body)
	}
	// The challenge is single-use.
	r.clk.step()
	if again := r.verify(ch, auth.MfaVerifyDto{Code: r.code(secret)}); again.code != 401 || again.body["code"] != "mfa_expired" {
		t.Fatalf("a used challenge was accepted: %d %v", again.code, again.body)
	}
}

func TestMfa_RecoveryCodesWorkOnceAndChallengesExpire(t *testing.T) {
	r := newMfaRig(t)
	ada, ws := r.person("ada")
	secret, codes := r.enroll(ada, ws)

	ok := r.verify(challengeOf(t, r.login(ada, "")), auth.MfaVerifyDto{RecoveryCode: codes[0]})
	if ok.code != 200 || ok.data()["mfaMethod"] != "recovery" {
		t.Fatalf("recovery code: %d %v", ok.code, ok.body)
	}
	if again := r.verify(challengeOf(t, r.login(ada, "")), auth.MfaVerifyDto{RecoveryCode: codes[0]}); again.code != 401 {
		t.Fatalf("a recovery code worked twice: %d", again.code)
	}
	st, _ := r.svc.Status(context.Background(), ada.Id)
	if st.RecoveryLeft != mfa.RecoveryCount-1 || st.RecoveryTotal != mfa.RecoveryCount {
		t.Fatalf("status: %+v", st)
	}

	// Five wrong attempts and the challenge is spent, even for the right code.
	ch := challengeOf(t, r.login(ada, ""))
	for i := 0; i < 5; i++ {
		r.verify(ch, auth.MfaVerifyDto{Code: "000000"})
	}
	if locked := r.verify(ch, auth.MfaVerifyDto{Code: r.code(secret)}); locked.code != 429 || locked.body["code"] != "mfa_locked" {
		t.Fatalf("after five attempts: %d %v", locked.code, locked.body)
	}

	// A challenge lasts ten minutes.
	ch = challengeOf(t, r.login(ada, ""))
	r.clk.t = r.clk.t.Add(11 * time.Minute)
	if late := r.verify(ch, auth.MfaVerifyDto{Code: r.code(secret)}); late.code != 401 || late.body["code"] != "mfa_expired" {
		t.Fatalf("an expired challenge: %d %v", late.code, late.body)
	}
	if none := r.verify("not-a-challenge", auth.MfaVerifyDto{Code: r.code(secret)}); none.code != 401 {
		t.Fatalf("an unknown challenge: %d", none.code)
	}
}

func TestMfa_ChangesNeedAFreshSecondFactor(t *testing.T) {
	r := newMfaRig(t)
	ada, ws := r.person("ada")
	secret, _ := r.enroll(ada, ws)
	st, _ := r.svc.Status(context.Background(), ada.Id)
	ctx := ctxOf(ada, ws)

	// No proof, or a proof for someone else, removes nothing.
	if out := decode(t, r.ctl.MfaRemove(&auth.MfaProofDto{Ctx: ctx, Id: st.TOTPFactorID})); out.code != 403 {
		t.Fatalf("remove without proof: %d %v", out.code, out.body)
	}
	bob, bws := r.person("bob")
	_, _ = r.enroll(bob, bws)
	bobStep := decode(t, r.ctl.MfaStepUp(&auth.MfaDto{Ctx: ctxOf(bob, bws)})).data()["challenge"].(string)
	if out := decode(t, r.ctl.MfaRemove(&auth.MfaProofDto{Ctx: ctx, Id: st.TOTPFactorID, Challenge: bobStep, Code: r.code(secret)})); out.code != 401 {
		t.Fatalf("another person's challenge: %d %v", out.code, out.body)
	}
	// Starting a new authenticator app once enrolled needs a proof too.
	if out := decode(t, r.ctl.MfaTotpStart(&auth.MfaProofDto{Ctx: ctx})); out.code != 403 {
		t.Fatalf("reconfigure without proof: %d %v", out.code, out.body)
	}

	// New recovery codes with a proof; the old ones stop working.
	step := decode(t, r.ctl.MfaStepUp(&auth.MfaDto{Ctx: ctx}))
	regen := decode(t, r.ctl.MfaRecovery(&auth.MfaProofDto{Ctx: ctx, Challenge: step.data()["challenge"].(string), Code: r.code(secret)}))
	if regen.code != 200 || len(regen.data()["recoveryCodes"].([]any)) != mfa.RecoveryCount {
		t.Fatalf("regenerate: %d %v", regen.code, regen.body)
	}
	r.clk.step()

	// Removing the only factor turns 2FA off and drops the codes.
	step = decode(t, r.ctl.MfaStepUp(&auth.MfaDto{Ctx: ctx}))
	out := decode(t, r.ctl.MfaRemove(&auth.MfaProofDto{Ctx: ctx, Id: st.TOTPFactorID, Challenge: step.data()["challenge"].(string), Code: r.code(secret)}))
	if out.code != 200 {
		t.Fatalf("remove: %d %v", out.code, out.body)
	}
	st, _ = r.svc.Status(context.Background(), ada.Id)
	if st.Enabled || st.RecoveryTotal != 0 {
		t.Fatalf("after removing the last factor: %+v", st)
	}
	if plain := r.login(ada, ""); plain.data()["accessToken"] == nil {
		t.Fatalf("sign-in without 2FA again: %v", plain.body)
	}
}

func (r *rig) join(u *user.User, ws, seat string) {
	r.t.Helper()
	link := &identityucw.UserClientWorkspace{UserId: u.Id, ClientId: r.client, WorkspaceId: ws, Status: &seat}
	link.Id = uuid.NewString()
	if err := r.db.Create(link).Error; err != nil {
		r.t.Fatal(err)
	}
}

func TestMfa_WorkspacePolicy(t *testing.T) {
	r := newMfaRig(t)
	owner, ws := r.person("owner")
	admin, _ := r.person("admin")
	dev, _ := r.person("dev")
	r.join(admin, ws, "admin")
	r.join(dev, ws, "member")

	var mu sync.Mutex
	var events []mfa.Event
	mfa.SetObserver(func(_ context.Context, e mfa.Event) { mu.Lock(); events = append(events, e); mu.Unlock() })
	defer mfa.SetObserver(nil)

	// Members cannot read or set it; an admin without 2FA cannot require it.
	if out := decode(t, r.ctl.MfaPolicy(&auth.MfaDto{Ctx: ctxOf(dev, ws)})); out.code != 403 {
		t.Fatalf("member reads policy: %d", out.code)
	}
	if out := decode(t, r.ctl.MfaSetPolicy(&auth.MfaPolicyDto{Ctx: ctxOf(admin, ws), Require: true})); out.code != 409 {
		t.Fatalf("admin without 2FA requires it: %d %v", out.code, out.body)
	}

	ownerSecret, _ := r.enroll(owner, ws)
	set := decode(t, r.ctl.MfaSetPolicy(&auth.MfaPolicyDto{Ctx: ctxOf(owner, ws), Require: true}))
	if set.code != 200 || set.data()["require"] != true {
		t.Fatalf("owner requires 2FA: %d %v", set.code, set.body)
	}
	missing, _ := set.data()["missing"].([]any)
	if len(missing) != 2 {
		t.Fatalf("two members have not set up 2FA: %v", missing)
	}

	// The session says so; the host's gate refuses until the member enrolls.
	links, _ := r.login(dev, "").data()["workspaces"].([]any)
	found := false
	for _, l := range links {
		if m := l.(map[string]any); m["workspaceId"] == ws {
			found = m["mfaRequired"] == true
		}
	}
	if !found {
		t.Fatalf("the workspace link carries mfaRequired: %v", links)
	}
	if err := r.svc.Gate(context.Background(), dev.Id, ws); !errors.Is(err, mfa.ErrEnrollmentRequired) {
		t.Fatalf("gate for a member without 2FA: %v", err)
	}
	_, _ = r.enroll(dev, ws)
	if err := r.svc.Gate(context.Background(), dev.Id, ws); err != nil {
		t.Fatalf("gate after enrolling: %v", err)
	}
	_, devWs := r.person("dev-elsewhere")
	if err := r.svc.Gate(context.Background(), dev.Id, devWs); err != nil {
		t.Fatalf("a workspace without the policy: %v", err)
	}

	// While it is required, the owner's only factor cannot be removed.
	st, _ := r.svc.Status(context.Background(), owner.Id)
	step := decode(t, r.ctl.MfaStepUp(&auth.MfaDto{Ctx: ctxOf(owner, ws)}))
	out := decode(t, r.ctl.MfaRemove(&auth.MfaProofDto{Ctx: ctxOf(owner, ws), Id: st.TOTPFactorID, Challenge: step.data()["challenge"].(string), Code: r.code(ownerSecret)}))
	if out.code != 409 {
		t.Fatalf("remove the last factor under the policy: %d %v", out.code, out.body)
	}

	mu.Lock()
	defer mu.Unlock()
	seen := map[string]int{}
	for _, e := range events {
		seen[e.Action]++
	}
	if seen["mfa.policy.change"] != 1 || seen["mfa.enable"] != 2 || seen["mfa.factor.add"] != 2 {
		t.Fatalf("events: %v", seen)
	}
}

func TestMfa_EverySignInPathAsksForTheSecondFactor(t *testing.T) {
	r := newMfaRig(t)
	r.ctl.SigninForTest().Magic = "424242"
	t.Setenv("NODE_ENV", "development")
	ada, ws := r.person("ada")
	_, _ = r.enroll(ada, ws)

	// An emailed code is a first factor only.
	challengeOf(t, r.otpVerify(*ada.Ref, "424242"))

	// The one-step service sign-in refuses outright.
	svc := pkg.NewAuthServiceForTest(r.db)
	if _, _, err := svc.Login(*ada.Ref, testPassword, r.client, ""); !errors.Is(err, pkg.ErrSecondFactorRequired) {
		t.Fatalf("one-step sign-in with 2FA: %v", err)
	}
	bob, _ := r.person("bob")
	if _, _, err := svc.Login(*bob.Ref, testPassword, r.client, ""); err != nil {
		t.Fatalf("one-step sign-in without 2FA: %v", err)
	}

	// The device's account chooser has nothing for a sign-in that stopped at the second step.
	bobDevice := deviceFrom(t, r.login(bob, ""))
	challengeOf(t, r.login(ada, cookieFor(bobDevice)))
	for _, s := range r.sessions(cookieFor(bobDevice)) {
		if u, _ := s["user"].(map[string]any); u["id"] == ada.Id {
			t.Fatalf("the device offers an account whose sign-in stopped at the second step: %v", s)
		}
	}

	// Security keys are off when no origin is configured.
	if out := decode(t, r.ctl.MfaKeyStart(&auth.MfaProofDto{Ctx: ctxOf(bob, "")})); out.code != 501 {
		t.Fatalf("keys without a relying party: %d %v", out.code, out.body)
	}
}

func TestChangeSecret(t *testing.T) {
	r := newMfaRig(t)
	ada, ws := r.person("ada")
	ctx := ctxOf(ada, ws)
	for _, c := range []struct {
		current, secret, code string
	}{
		{"wrong password", "new-secret-1", "secret_mismatch"},
		{testPassword, "short", "secret_length"},
		{testPassword, "this one is far too long", "secret_length"},
	} {
		out := decode(t, r.ctl.ChangeSecret(&auth.ChangeSecretDto{Ctx: ctx, Current: c.current, Secret: c.secret}))
		if out.code != 400 || out.body["code"] != c.code {
			t.Fatalf("%q → %q: %d %v", c.current, c.secret, out.code, out.body)
		}
	}
	if out := decode(t, r.ctl.ChangeSecret(&auth.ChangeSecretDto{Ctx: ctx, Current: testPassword, Secret: "new-secret-1"})); out.code != 200 {
		t.Fatalf("change: %d %v", out.code, out.body)
	}
	if old := r.login(ada, ""); old.code != 401 {
		t.Fatalf("the old password still signs in: %d", old.code)
	}
	email := *ada.Ref
	if fresh := decode(t, r.ctl.Login(&auth.LoginDto{Ref: email, Secret: "new-secret-1", ClientId: r.client})); fresh.code != 200 {
		t.Fatalf("the new password: %d %v", fresh.code, fresh.body)
	}
}
