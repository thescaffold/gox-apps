package tests

import (
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/awesome-goose/goose/modules/sql"
	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	"github.com/thescaffold/gox-apps/libs/identity/app/user"
	"github.com/thescaffold/gox-packages/libs/core/events"
	"github.com/thescaffold/gox-packages/libs/core/events/dbevents"
)

type created struct {
	mu  sync.Mutex
	got []map[string]any
}

func (c *created) add(_ string, p any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	m, _ := p.(map[string]any)
	c.got = append(c.got, m)
}

func (c *created) all() []map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]map[string]any(nil), c.got...)
}

// A new account is announced once, with the workspace its sign-up made; signing
// in again announces nothing (PLAN U12: hosts provision eagerly from it).
func TestNewAccountsAreAnnounced(t *testing.T) {
	r := plRig(t)
	bus := events.NewBus()
	heard := &created{}
	bus.Subscribe(user.EventUserCreated, heard.add)
	r.ctl.SetTrackerForTest(events.NewTrackerService(bus))

	addr := uniq("announce")
	first := data(r.otpVerify(addr, "424242"))
	if first["isNew"] != true {
		t.Fatalf("%v", first)
	}
	_ = data(r.otpVerify(addr, "424242"))
	got := heard.all()
	if len(got) != 1 {
		t.Fatalf("%d announcements for one new account, want 1: %v", len(got), got)
	}
	id := first["user"].(map[string]any)["id"]
	ws, _ := first["workspaces"].([]any)
	if got[0]["userId"] != id || got[0]["clientId"] != r.client || got[0]["method"] != "passwordless" || got[0]["email"] != addr {
		t.Fatalf("announcement %v", got[0])
	}
	if len(ws) != 1 || got[0]["workspaceId"] == "" {
		t.Fatalf("the announcement lacks the new workspace: %v / %v", got[0], ws)
	}

	// The password sign-up path announces too.
	rid := uuid.NewString()
	email := "announce-" + rid[:8] + "@example.test"
	token := r.ctl.SeedRegistrationForTest(rid, email, r.client)
	rep := decode(t, r.ctl.Secret(&auth.SecretDto{Token: token, Secret: "a long enough secret", ConfirmSecret: "a long enough secret", FirstName: "Ann", LastName: "Ounce"}))
	if rep.code != 200 {
		t.Fatalf("secret: %d %v", rep.code, rep.body)
	}
	got = heard.all()
	if len(got) != 2 || got[1]["email"] != email || got[1]["method"] != "password" || got[1]["workspaceId"] == "" {
		t.Fatalf("password sign-up announcement: %v", got)
	}
}

// With core's dbevents on, the insert itself is an apps.db.identityusers.after-insert
// event, which identity's own Subscriptions (and any host) listen for.
func TestUserInsertIsADBEvent(t *testing.T) {
	r := newRig(t)
	bus := events.NewBus()
	heard := &created{}
	bus.Subscribe(dbevents.Name("IdentityUsers", dbevents.AfterInsert), heard.add)
	stop := dbevents.Publish(bus)

	email := "dbevent-" + uuid.NewString()[:8] + "@example.test"
	typ := "email"
	u := &user.User{Name: "Db Event", Email: email, Ref: &email, Type: &typ}
	users := &user.UserEntity{Entity: (&sql.Entity[user.User]{}).With(r.q())}
	users.OnRegister()
	if err := users.Insert(u); err != nil {
		t.Fatal(err)
	}
	stop() // waits for delivery

	got := heard.all()
	if len(got) != 1 {
		t.Fatalf("%d after-insert events, want 1", len(got))
	}
	meta, _ := got[0]["meta"].(map[string]any)
	if got[0]["entityId"] != u.Id || meta["email"] != email || got[0]["action"] != "create" {
		t.Fatalf("event %v", got[0])
	}
	if _, leaked := meta["secret"]; leaked {
		t.Fatal("the password hash reached the event")
	}
}
