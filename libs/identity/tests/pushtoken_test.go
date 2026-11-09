package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	apppush "github.com/thescaffold/gox-apps/libs/identity/app/pushtoken"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/pushtoken"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// PLAN U5: an install registers its push token; it is refreshed in place,
// moves when someone else signs in on that install, and goes on sign-out.
func TestPushTokens_RegisterRefreshMoveAndDeregister(t *testing.T) {
	r := newRig(t)
	svc := pushtoken.New(r.db)
	ctl := apppush.NewControllerForTest(svc)
	ada, _ := r.person("ada")
	lin, _ := r.person("lin")
	tok := "fcm-" + uuid.NewString()

	rep := decode(t, ctl.Register(&apppush.RegisterReq{Ctx: ntxctx.NTXContext{UserID: ada.Id}, Platform: "Android", Token: tok, AppVersion: "1.2.0", Locale: "en-NG"}))
	if rep.code != 200 {
		t.Fatalf("register: %d %v", rep.code, rep.body)
	}
	d := data(rep)
	if d["platform"] != "android" || d["provider"] != "fcm" || d["appVersion"] != "1.2.0" || d["locale"] != "en-NG" {
		t.Fatalf("registration = %v", d)
	}
	if _, echoed := d["token"]; echoed {
		t.Fatal("the token was echoed back")
	}
	// Again with a new app version: refreshed, not duplicated.
	decode(t, ctl.Register(&apppush.RegisterReq{Ctx: ntxctx.NTXContext{UserID: ada.Id}, Platform: "android", Token: tok, AppVersion: "1.3.0", Locale: "fr"}))
	got, _ := svc.For(context.Background(), ada.Id)
	if len(got) != 1 || got[0].AppVersion != "1.3.0" || got[0].Locale != "fr" {
		t.Fatalf("after refresh: %+v", got)
	}
	// Lin signs in on the same install: the token is Lin's now.
	decode(t, ctl.Register(&apppush.RegisterReq{Ctx: ntxctx.NTXContext{UserID: lin.Id}, Platform: "android", Token: tok}))
	if a, _ := svc.For(context.Background(), ada.Id); len(a) != 0 {
		t.Fatalf("ada still has the moved token: %+v", a)
	}
	if l, _ := svc.For(context.Background(), lin.Id); len(l) != 1 {
		t.Fatalf("lin's tokens: %+v", l)
	}
	// Ada cannot remove Lin's registration.
	if rem := data(decode(t, ctl.Deregister(&apppush.DeregisterReq{Ctx: ntxctx.NTXContext{UserID: ada.Id}, Token: tok}))); rem["removed"] != false {
		t.Fatalf("ada removed lin's token: %v", rem)
	}
	if rem := data(decode(t, ctl.Deregister(&apppush.DeregisterReq{Ctx: ntxctx.NTXContext{UserID: lin.Id}, Token: tok}))); rem["removed"] != true {
		t.Fatalf("lin could not remove their token: %v", rem)
	}

	// Refusals.
	for name, req := range map[string]*apppush.RegisterReq{
		"anonymous":    {Platform: "ios", Token: "x"},
		"bad platform": {Ctx: ntxctx.NTXContext{UserID: ada.Id}, Platform: "symbian", Token: "x"},
		"no token":     {Ctx: ntxctx.NTXContext{UserID: ada.Id}, Platform: "ios"},
	} {
		if rep := decode(t, ctl.Register(req)); rep.code == 200 {
			t.Fatalf("%s: accepted", name)
		}
	}
	if _, err := svc.Register(context.Background(), "", pushtoken.Registration{Platform: "ios", Token: "x"}); !errors.Is(err, pushtoken.ErrNoUser) {
		t.Fatal(err)
	}
}

// Signing a device out removes the tokens registered for it; signing out
// everywhere removes all of the person's.
func TestPushTokens_GoWithSignOut(t *testing.T) {
	r := newRig(t)
	svc := pushtoken.New(r.db)
	ada, _ := r.person("ada")
	device := deviceFrom(t, r.login(ada, ""))
	ctx := context.Background()
	if _, err := svc.Register(ctx, ada.Id, pushtoken.Registration{Platform: "ios", Token: "apns-" + uuid.NewString(), DeviceId: device}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Register(ctx, ada.Id, pushtoken.Registration{Platform: "android", Token: "fcm-" + uuid.NewString()}); err != nil {
		t.Fatal(err)
	}
	decode(t, r.ctl.Logout(&auth.LogoutDto{Cookie: cookieFor(device), Device: "yes", ClientId: r.client}))
	left, _ := svc.For(ctx, ada.Id)
	if len(left) != 1 || left[0].Platform != "android" {
		t.Fatalf("after signing the device out: %+v", left)
	}
	decode(t, r.ctl.Logout(&auth.LogoutDto{Ctx: ntxctx.NTXContext{UserID: ada.Id}, ClientId: r.client}))
	if left, _ = svc.For(ctx, ada.Id); len(left) != 0 {
		t.Fatalf("after signing out everywhere: %+v", left)
	}
}

// A native app signs out with its refresh token and names its push token.
func TestPushTokens_RefreshTokenSignOutTakesThePushToken(t *testing.T) {
	r := newRig(t)
	svc := pushtoken.New(r.db)
	ada, _ := r.person("ada")
	phone := refreshTokenOf(t, r.login(ada, ""))
	ctx := context.Background()
	tok := "apns-" + uuid.NewString()
	other := "fcm-" + uuid.NewString()
	for _, x := range []string{tok, other} {
		if _, err := svc.Register(ctx, ada.Id, pushtoken.Registration{Platform: "ios", Token: x}); err != nil {
			t.Fatal(err)
		}
	}
	if rep := decode(t, r.ctl.Logout(&auth.LogoutDto{RefreshToken: phone, PushToken: tok})); rep.code != 200 {
		t.Fatalf("sign-out: %d %v", rep.code, rep.body)
	}
	left, _ := svc.For(ctx, ada.Id)
	if len(left) != 1 {
		t.Fatalf("after signing the phone out: %+v", left)
	}
	// A refused sign-out removes nothing.
	decode(t, r.ctl.Logout(&auth.LogoutDto{RefreshToken: "not-a-token", PushToken: other}))
	if left, _ = svc.For(ctx, ada.Id); len(left) != 1 {
		t.Fatalf("a refused sign-out removed a token: %+v", left)
	}
}
