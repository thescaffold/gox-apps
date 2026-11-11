package tests

import (
	"strings"
	"testing"
	"time"

	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/events"
)

// A person changes their sign-in email by proving the new address: a code goes to it, the change
// happens only when the code is entered, and the old address is told.
func TestEmailChangeIsVerified(t *testing.T) {
	r := newRig(t)
	r.db.Exec(`TRUNCATE "IdentityOtps"`)
	t.Setenv("NODE_ENV", "development")
	t.Setenv("IDENTITY_MAGIC_OTP", "")
	bus := events.NewBus()
	heard := &created{}
	bus.Subscribe("apps.notification.message.new", heard.add)
	r.ctl.SetTrackerForTest(events.NewTrackerService(bus))

	u, _ := r.person("mover")
	taken, _ := r.person("taken")
	old := u.Email
	me := ntxctx.NTXContext{UserID: u.Id, ClientID: r.client}
	next := uniq("moved")

	if rep := decode(t, r.ctl.EmailChangeStart(&auth.EmailChangeDto{Ctx: me, Email: "not an address"})); rep.code != 400 {
		t.Fatalf("a bad address: %d", rep.code)
	}
	if rep := decode(t, r.ctl.EmailChangeStart(&auth.EmailChangeDto{Ctx: me, Email: taken.Email})); rep.code != 409 {
		t.Fatalf("someone else's address: %d", rep.code)
	}
	if rep := decode(t, r.ctl.EmailChangeStart(&auth.EmailChangeDto{Email: next})); rep.code != 401 {
		t.Fatalf("without a session: %d", rep.code)
	}
	if rep := decode(t, r.ctl.EmailChangeStart(&auth.EmailChangeDto{Ctx: me, Email: strings.ToUpper(next)})); rep.code != 200 {
		t.Fatalf("start: %d %v", rep.code, rep.body)
	}
	code := ""
	waitFor(t, func() bool {
		for _, m := range heard.all() {
			if d, _ := m["data"].(map[string]any); m["key"] == "email-change" && d["owner"] == next {
				code, _ = d["code"].(string)
			}
		}
		return code != ""
	})

	// Nothing changes until the code is entered; a wrong code changes nothing either.
	var email string
	r.db.Raw(`SELECT email FROM "IdentityUsers" WHERE id = ?`, u.Id).Scan(&email)
	if email != old {
		t.Fatalf("the email changed before the code: %q", email)
	}
	wrong := "000000"
	if code == wrong {
		wrong = "111111"
	}
	if rep := decode(t, r.ctl.EmailChangeVerify(&auth.EmailChangeDto{Ctx: me, Email: next, Code: wrong})); rep.code != 401 {
		t.Fatalf("a wrong code: %d", rep.code)
	}
	// Another person cannot use the code.
	if rep := decode(t, r.ctl.EmailChangeVerify(&auth.EmailChangeDto{Ctx: ntxctx.NTXContext{UserID: taken.Id, ClientID: r.client}, Email: next, Code: code})); rep.code != 401 {
		t.Fatalf("someone else's code: %d", rep.code)
	}
	rep := decode(t, r.ctl.EmailChangeVerify(&auth.EmailChangeDto{Ctx: me, Email: next, Code: code}))
	if rep.code != 200 || rep.data()["email"] != next {
		t.Fatalf("verify: %d %v", rep.code, rep.body)
	}
	var ref string
	r.db.Raw(`SELECT email, ref FROM "IdentityUsers" WHERE id = ?`, u.Id).Row().Scan(&email, &ref)
	if email != next || ref != next {
		t.Fatalf("stored %q %q, want %q", email, ref, next)
	}
	// The code is spent.
	if rep := decode(t, r.ctl.EmailChangeVerify(&auth.EmailChangeDto{Ctx: me, Email: next, Code: code})); rep.code != 401 {
		t.Fatalf("the code twice: %d", rep.code)
	}
	waitFor(t, func() bool {
		for _, m := range heard.all() {
			if d, _ := m["data"].(map[string]any); m["key"] == "email-changed" && d["owner"] == old && d["ref"] == next {
				return true
			}
		}
		return false
	})
	// The new address signs in; the old one does not.
	u.Ref = &next
	if out := r.login(u, ""); out.code != 200 {
		t.Fatalf("sign in with the new address: %d %v", out.code, out.body)
	}
	if out := decode(t, r.ctl.Login(&auth.LoginDto{Ref: old, Secret: testPassword, ClientId: r.client})); out.code == 200 {
		t.Fatal("the old address still signs in")
	}
}

func waitFor(t *testing.T, ok func() bool) {
	t.Helper()
	for i := 0; i < 100; i++ {
		if ok() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("timed out")
}
