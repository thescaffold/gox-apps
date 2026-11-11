package app

import (
	"net/http"
	"testing"

	notificationprovider "github.com/thescaffold/gox-apps/libs/notification/pkg/provider"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

type fakeIdentity map[string]map[string]any

func (f fakeIdentity) FetchByUserId(id string) (*notificationprovider.UserContext, error) {
	if u, ok := f[id]; ok {
		return &notificationprovider.UserContext{User: u}, nil
	}
	return nil, nil
}
func (f fakeIdentity) FetchByOwner(string) (*notificationprovider.UserContext, error) {
	return nil, nil
}

// A subscription is keyed by an address; only its owner or a platform admin reaches it (U26).
func TestSubscriptionIsTheAddressOwnersOnly(t *testing.T) {
	t.Setenv("ROOT_USER_ID", "root-1")
	p := &notificationprovider.ProviderService{}
	p.SetIdentity(fakeIdentity{"alice": {"ref": "alice@example.com", "email": "alice@example.com"}})
	c := &AppController{provider: p}

	alice := ntxctx.NTXContext{UserID: "alice"}
	bob := ntxctx.NTXContext{UserID: "bob"}
	if !c.ownsAddress(alice, "alice@example.com") {
		t.Error("alice cannot reach her own subscription")
	}
	if c.ownsAddress(bob, "alice@example.com") {
		t.Error("bob reached alice's subscription")
	}
	if !c.ownsAddress(ntxctx.NTXContext{UserID: "root-1"}, "alice@example.com") {
		t.Error("a platform admin is refused")
	}
	if out := c.GetSubscription(&GetSubscriptionDto{Ref: "alice@example.com", Ctx: bob}); out.Code() != http.StatusForbidden {
		t.Errorf("GET as bob = %d, want 403", out.Code())
	}
	if out := c.UpdateSubscription(&UpdateSubscriptionDto{Ref: "alice@example.com", Rules: []byte(`{}`), Ctx: bob}); out.Code() != http.StatusForbidden {
		t.Errorf("PATCH as bob = %d, want 403", out.Code())
	}
	// no IdentityFetcher wired: only an admin
	if (&AppController{provider: &notificationprovider.ProviderService{}}).ownsAddress(alice, "alice@example.com") {
		t.Error("without an IdentityFetcher a non-admin must be refused")
	}
}
