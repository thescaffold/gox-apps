package tests

import (
	"errors"
	"testing"

	"github.com/thescaffold/gox-apps/libs/identity/pkg/org"
)

// Seat changes follow one set of rules: only owners and admins change seats,
// nobody grants or touches a seat above their own, only an owner grants or
// removes owner, and a workspace keeps at least one owner.
func TestOrg_SeatRules(t *testing.T) {
	e := newOrgEnv(t)
	ws := newWS()
	owner := e.seat(t, ws, "Olive Owner", "owner")
	admin := e.seat(t, ws, "Adam Admin", "admin")
	member := e.seat(t, ws, "Mia Member", "member")
	guest := e.seat(t, ws, "Gus Guest", "guest")
	outsider := e.seat(t, newWS(), "Otto Outsider", "owner")

	seat := func(u string) string {
		s, err := e.svc.Seat(e.ctx, ws, u)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}

	if err := e.svc.SetSeat(e.ctx, ws, member, guest, "member"); !errors.Is(err, org.ErrNotAllowed) {
		t.Fatalf("a member changing seats: %v", err)
	}
	if err := e.svc.SetSeat(e.ctx, ws, admin, member, "owner"); !errors.Is(err, org.ErrNotAllowed) {
		t.Fatalf("an admin granting owner: %v", err)
	}
	if err := e.svc.SetSeat(e.ctx, ws, admin, owner, "member"); !errors.Is(err, org.ErrNotAllowed) {
		t.Fatalf("an admin demoting the owner: %v", err)
	}
	if err := e.svc.SetSeat(e.ctx, ws, admin, member, "superuser"); !errors.Is(err, org.ErrInvalid) {
		t.Fatalf("an unknown seat: %v", err)
	}
	if err := e.svc.SetSeat(e.ctx, ws, admin, outsider, "member"); !errors.Is(err, org.ErrNotMember) {
		t.Fatalf("someone from another workspace: %v", err)
	}
	if err := e.svc.SetSeat(e.ctx, ws, owner, owner, "admin"); !errors.Is(err, org.ErrInvalid) {
		t.Fatalf("the last owner stepping down: %v", err)
	}

	if err := e.svc.SetSeat(e.ctx, ws, admin, guest, "member"); err != nil || seat(guest) != "member" {
		t.Fatalf("an admin promoting a guest: %v, seat %q", err, seat(guest))
	}
	if err := e.svc.SetSeat(e.ctx, ws, owner, admin, "owner"); err != nil || seat(admin) != "owner" {
		t.Fatalf("an owner naming a second owner: %v", err)
	}
	// With two owners, one may step down.
	if err := e.svc.SetSeat(e.ctx, ws, owner, owner, "admin"); err != nil || seat(owner) != "admin" {
		t.Fatalf("an owner stepping down with another owner left: %v", err)
	}
	if ok, _ := e.svc.IsAdmin(e.ctx, ws, owner); !ok {
		t.Fatal("an admin seat must still be an admin")
	}

	if err := e.svc.RemoveSeat(e.ctx, ws, owner, member); err != nil || seat(member) != "" {
		t.Fatalf("an admin removing a member: %v, seat %q", err, seat(member))
	}
	if err := e.svc.RemoveSeat(e.ctx, ws, owner, admin); !errors.Is(err, org.ErrNotAllowed) {
		t.Fatalf("an admin removing the (now) owner: %v", err)
	}
}

// A product can register a richer, ordered seat list; the defaults are
// restored for the other tests.
func TestOrg_CustomSeatRoles(t *testing.T) {
	defaults := org.SeatRoles()
	t.Cleanup(func() { _ = org.SetSeatRoles(defaults) })

	if err := org.SetSeatRoles([]string{"admin", "owner"}); !errors.Is(err, org.ErrInvalid) {
		t.Fatalf("owner must come first: %v", err)
	}
	if err := org.SetSeatRoles([]string{"owner", "admin", "maintainer", "developer", "member", "reporter", "guest"}); err != nil {
		t.Fatal(err)
	}

	e := newOrgEnv(t)
	ws := newWS()
	admin := e.seat(t, ws, "Ann Admin", "admin")
	dev := e.seat(t, ws, "Dee Dev", "developer")
	if err := e.svc.SetSeat(e.ctx, ws, admin, dev, "maintainer"); err != nil {
		t.Fatalf("granting a registered seat: %v", err)
	}
	if s, _ := e.svc.Seat(e.ctx, ws, dev); s != "maintainer" {
		t.Fatalf("seat = %q", s)
	}
}
