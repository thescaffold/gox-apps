package tests

import (
	"testing"

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

func data(rep reply) map[string]any {
	d, _ := rep.body["data"].(map[string]any)
	return d
}

func TestPasswordless_FirstSignInCreatesTheAccountAndLaterOnesMatchIt(t *testing.T) {
	r := plRig(t)
	first := r.otpVerify("Ada@Example.com", "424242")
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

	second := data(r.otpVerify("ada@example.com", "424242")) // the same address, different case
	if second["isNew"] != false || second["user"].(map[string]any)["id"] != user1 {
		t.Fatalf("a second sign-in must match the same user: %v", second)
	}
	if ws, _ := second["workspaces"].([]any); len(ws) != 1 {
		t.Fatalf("no second workspace: %v", second["workspaces"])
	}
	var n int64
	r.db.Raw(`SELECT COUNT(*) FROM "IdentityUsers" WHERE email = 'ada@example.com'`).Scan(&n)
	if n != 1 {
		t.Fatalf("%d users for one address", n)
	}
}

func TestPasswordless_PhoneNumbersNormaliseToOneAccount(t *testing.T) {
	r := plRig(t)
	a := data(r.otpVerify("08012345678", "424242"))
	b := data(r.otpVerify("+234 801 234 5678", "424242"))
	if a["isNew"] != true || b["isNew"] != false || a["user"].(map[string]any)["id"] != b["user"].(map[string]any)["id"] {
		t.Fatalf("%v %v", a, b)
	}
	var phone string
	r.db.Raw(`SELECT phone FROM "IdentityUsers" WHERE ref = '+2348012345678'`).Scan(&phone)
	if phone != "+2348012345678" {
		t.Fatalf("phone stored as %q", phone)
	}
}

func TestPasswordless_WrongCodesAreRefusedLockedAndTheAccountIsNotCreated(t *testing.T) {
	r := plRig(t)
	r.ctl.SigninForTest().Magic = ""
	req := decode(t, r.ctl.OtpRequest(&auth.OtpRequestDto{Ref: "lin@example.com", ClientId: r.client}))
	if req.code != 200 {
		t.Fatalf("%d %v", req.code, req.body)
	}
	for i := 0; i < 5; i++ {
		if rep := r.otpVerify("lin@example.com", "000000"); rep.code != 401 {
			t.Fatalf("guess %d: %d", i, rep.code)
		}
	}
	if rep := r.otpVerify("lin@example.com", "000000"); rep.code != 429 {
		t.Fatalf("a locked code: %d", rep.code)
	}
	var n int64
	r.db.Raw(`SELECT COUNT(*) FROM "IdentityUsers" WHERE email = 'lin@example.com'`).Scan(&n)
	if n != 0 {
		t.Fatal("an account was created without a valid code")
	}
}

func TestPasswordless_RequestsAreThrottledAndInputIsValidated(t *testing.T) {
	r := plRig(t)
	if rep := decode(t, r.ctl.OtpRequest(&auth.OtpRequestDto{Ref: "x@example.com", ClientId: r.client})); rep.code != 200 {
		t.Fatalf("%d", rep.code)
	}
	if rep := decode(t, r.ctl.OtpRequest(&auth.OtpRequestDto{Ref: "x@example.com", ClientId: r.client})); rep.code != 429 {
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
	r.otpVerify("known@example.com", "424242") // creates the account
	known := decode(t, r.ctl.OtpRequest(&auth.OtpRequestDto{Ref: "known@example.com", ClientId: r.client}))
	unknown := decode(t, r.ctl.OtpRequest(&auth.OtpRequestDto{Ref: "stranger@example.com", ClientId: r.client}))
	if known.code != unknown.code || known.body["message"] != unknown.body["message"] {
		t.Fatalf("existing and unknown accounts must look the same: %v vs %v", known.body, unknown.body)
	}
}

func TestPasswordless_SuspendedAndClosedAccountsCannotSignIn(t *testing.T) {
	r := plRig(t)
	d := data(r.otpVerify("gone@example.com", "424242"))
	id := d["user"].(map[string]any)["id"].(string)
	r.db.Exec(`INSERT INTO "IdentityAttributes" (id, created_at, updated_at, user_id, key, value) VALUES (gen_random_uuid()::text, now(), now(), ?, 'suspended_at', 'x')`, id)
	if rep := r.otpVerify("gone@example.com", "424242"); rep.code != 401 {
		t.Fatalf("suspended: %d", rep.code)
	}
	r.db.Exec(`DELETE FROM "IdentityAttributes" WHERE user_id = ?`, id)
	r.db.Exec(`INSERT INTO "IdentityAttributes" (id, created_at, updated_at, user_id, key, value) VALUES (gen_random_uuid()::text, now(), now(), ?, 'closed_at', 'x')`, id)
	if rep := r.otpVerify("gone@example.com", "424242"); rep.code != 401 {
		t.Fatalf("closed: %d", rep.code)
	}
}

func TestPasswordless_AccountsHaveNoPasswordSoPasswordLoginCannotBeUsedToGetIn(t *testing.T) {
	r := plRig(t)
	r.otpVerify("nopass@example.com", "424242")
	for _, guess := range []string{"", "password", "424242", "nopass@example.com"} {
		if guess == "" {
			continue
		}
		rep := decode(t, r.ctl.Login(&auth.LoginDto{Ref: "nopass@example.com", Secret: guess, ClientId: r.client}))
		if rep.code != 401 {
			t.Fatalf("password %q signed in: %d", guess, rep.code)
		}
	}
}

func TestPasswordless_TheMagicCodeIsRefusedInProduction(t *testing.T) {
	r := plRig(t)
	t.Setenv("NODE_ENV", "production")
	if rep := r.otpVerify("prod@example.com", "424242"); rep.code != 401 {
		t.Fatalf("the sandbox code worked in production: %d", rep.code)
	}
}
