package tests

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
)

func plRig(t *testing.T) *rig {
	t.Helper()
	r := newRig(t)
	r.db.Exec(`TRUNCATE "IdentityOtps"`)
	r.ctl.SigninForTest().Magic = "424242"
	t.Setenv("NODE_ENV", "development")
	return r
}

func (r *rig) otpVerify(ref, code string) reply {
	r.t.Helper()
	return decode(r.t, r.ctl.OtpVerify(&auth.OtpVerifyDto{Ref: ref, Code: code, ClientId: r.client}))
}

// uniq makes a sign-in reference nobody used in an earlier run (the test
// database persists accounts).
func uniq(tag string) string { return fmt.Sprintf("%s-%s@example.com", tag, uuid.NewString()[:8]) }

func uniqPhone() (national, e164 string) {
	n := fmt.Sprintf("%07d", rand.Intn(10_000_000))
	return "0801" + n[:7], "+234801" + n[:7]
}

func data(rep reply) map[string]any {
	d, _ := rep.body["data"].(map[string]any)
	return d
}

func TestPasswordless_FirstSignInCreatesTheAccountAndLaterOnesMatchIt(t *testing.T) {
	r := plRig(t)
	addr := uniq("ada")
	first := r.otpVerify(strings.ToUpper(addr[:1])+addr[1:], "424242")
	if first.code != 200 {
		t.Fatalf("%d %v", first.code, first.body)
	}
	d := data(first)
	if d["isNew"] != true || d["accessToken"] == "" || d["refreshToken"] == "" {
		t.Fatalf("%v", d)
	}
	if ws, _ := d["workspaces"].([]any); len(ws) != 1 {
		t.Fatalf("a new account gets one workspace: %v", d["workspaces"])
	}
	user1 := d["user"].(map[string]any)["id"]

	second := data(r.otpVerify(addr, "424242")) // the same address, different case
	if second["isNew"] != false || second["user"].(map[string]any)["id"] != user1 {
		t.Fatalf("a second sign-in must match the same user: %v", second)
	}
	if ws, _ := second["workspaces"].([]any); len(ws) != 1 {
		t.Fatalf("no second workspace: %v", second["workspaces"])
	}
	var n int64
	r.db.Raw(`SELECT COUNT(*) FROM "IdentityUsers" WHERE email = ?`, addr).Scan(&n)
	if n != 1 {
		t.Fatalf("%d users for one address", n)
	}
}

func TestPasswordless_PhoneNumbersNormaliseToOneAccount(t *testing.T) {
	r := plRig(t)
	national, e164 := uniqPhone()
	spaced := e164[:4] + " " + e164[4:7] + " " + e164[7:10] + " " + e164[10:]
	a := data(r.otpVerify(national, "424242"))
	b := data(r.otpVerify(spaced, "424242"))
	if a["isNew"] != true || b["isNew"] != false || a["user"].(map[string]any)["id"] != b["user"].(map[string]any)["id"] {
		t.Fatalf("%v %v", a, b)
	}
	var phone string
	r.db.Raw(`SELECT phone FROM "IdentityUsers" WHERE ref = ?`, e164).Scan(&phone)
	if phone != e164 {
		t.Fatalf("phone stored as %q", phone)
	}
}

func TestPasswordless_WrongCodesAreRefusedLockedAndTheAccountIsNotCreated(t *testing.T) {
	r := plRig(t)
	r.ctl.SigninForTest().Magic = ""
	lin := uniq("lin")
	req := decode(t, r.ctl.OtpRequest(&auth.OtpRequestDto{Ref: lin, ClientId: r.client}))
	if req.code != 200 {
		t.Fatalf("%d %v", req.code, req.body)
	}
	for i := 0; i < 5; i++ {
		if rep := r.otpVerify(lin, "000000"); rep.code != 401 {
			t.Fatalf("guess %d: %d", i, rep.code)
		}
	}
	if rep := r.otpVerify(lin, "000000"); rep.code != 429 {
		t.Fatalf("a locked code: %d", rep.code)
	}
	var n int64
	r.db.Raw(`SELECT COUNT(*) FROM "IdentityUsers" WHERE email = ?`, lin).Scan(&n)
	if n != 0 {
		t.Fatal("an account was created without a valid code")
	}
}

func TestPasswordless_RequestsAreThrottledAndInputIsValidated(t *testing.T) {
	r := plRig(t)
	x := uniq("x")
	if rep := decode(t, r.ctl.OtpRequest(&auth.OtpRequestDto{Ref: x, ClientId: r.client})); rep.code != 200 {
		t.Fatalf("%d", rep.code)
	}
	if rep := decode(t, r.ctl.OtpRequest(&auth.OtpRequestDto{Ref: x, ClientId: r.client})); rep.code != 429 {
		t.Fatalf("a second request within the gap: %d", rep.code)
	}
	for _, bad := range []string{"nonsense", "12345", ""} {
		if rep := decode(t, r.ctl.OtpRequest(&auth.OtpRequestDto{Ref: bad, ClientId: r.client})); rep.code != 400 {
			t.Errorf("%q: %d", bad, rep.code)
		}
		if rep := r.otpVerify(bad, "424242"); rep.code != 400 {
			t.Errorf("verify %q: %d", bad, rep.code)
		}
	}
}

func TestPasswordless_TheRequestAnswerDoesNotRevealWhetherAnAccountExists(t *testing.T) {
	r := plRig(t)
	k, st := uniq("known"), uniq("stranger")
	r.otpVerify(k, "424242") // creates the account
	known := decode(t, r.ctl.OtpRequest(&auth.OtpRequestDto{Ref: k, ClientId: r.client}))
	unknown := decode(t, r.ctl.OtpRequest(&auth.OtpRequestDto{Ref: st, ClientId: r.client}))
	if known.code != unknown.code || known.body["message"] != unknown.body["message"] {
		t.Fatalf("existing and unknown accounts must look the same: %v vs %v", known.body, unknown.body)
	}
}

func TestPasswordless_SuspendedAndClosedAccountsCannotSignIn(t *testing.T) {
	r := plRig(t)
	gone := uniq("gone")
	d := data(r.otpVerify(gone, "424242"))
	id := d["user"].(map[string]any)["id"].(string)
	r.db.Exec(`INSERT INTO "IdentityAttributes" (id, created_at, updated_at, user_id, key, value) VALUES (gen_random_uuid()::text, now(), now(), ?, 'suspended_at', 'x')`, id)
	if rep := r.otpVerify(gone, "424242"); rep.code != 401 {
		t.Fatalf("suspended: %d", rep.code)
	}
	r.db.Exec(`DELETE FROM "IdentityAttributes" WHERE user_id = ?`, id)
	r.db.Exec(`INSERT INTO "IdentityAttributes" (id, created_at, updated_at, user_id, key, value) VALUES (gen_random_uuid()::text, now(), now(), ?, 'closed_at', 'x')`, id)
	if rep := r.otpVerify(gone, "424242"); rep.code != 401 {
		t.Fatalf("closed: %d", rep.code)
	}
}

func TestPasswordless_AccountsHaveNoPasswordSoPasswordLoginCannotBeUsedToGetIn(t *testing.T) {
	r := plRig(t)
	np := uniq("nopass")
	r.otpVerify(np, "424242")
	for _, guess := range []string{"", "password", "424242", np} {
		if guess == "" {
			continue
		}
		rep := decode(t, r.ctl.Login(&auth.LoginDto{Ref: np, Secret: guess, ClientId: r.client}))
		if rep.code != 401 {
			t.Fatalf("password %q signed in: %d", guess, rep.code)
		}
	}
}

func TestPasswordless_TheMagicCodeIsRefusedInProduction(t *testing.T) {
	r := plRig(t)
	t.Setenv("NODE_ENV", "production")
	if rep := r.otpVerify(uniq("prod"), "424242"); rep.code != 401 {
		t.Fatalf("the sandbox code worked in production: %d", rep.code)
	}
}
