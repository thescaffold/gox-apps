package tests

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/thescaffold/gox-apps/libs/identity/pkg/apikey"
	"gorm.io/gorm"
)

var kctx = context.Background()

func keySvc(t *testing.T) (*apikey.Service, *gorm.DB) {
	t.Helper()
	db := identityTestDB(t)
	db.Exec(`TRUNCATE "IdentityApiKeys"`)
	return &apikey.Service{DB: db, Brand: "lgr"}, db
}

func issue(t *testing.T, s *apikey.Service, client, ws, kind, env string) apikey.Key {
	t.Helper()
	k, err := s.Issue(kctx, "usr-1", client, ws, "server key", kind, env, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func TestApiKey_FormatShownOnceAndStoredOnlyAsAHash(t *testing.T) {
	s, db := keySvc(t)
	k := issue(t, s, "c1", "w1", apikey.Secret, apikey.Test)
	if !strings.HasPrefix(k.Secret, "lgr_test_sk_") || len(k.Secret) < 30 {
		t.Fatalf("%q", k.Secret)
	}
	if !strings.HasPrefix(k.Secret, k.Prefix) || len(k.Prefix) >= len(k.Secret) {
		t.Fatalf("prefix %q is not a safe display prefix of %q", k.Prefix, k.Secret)
	}
	pk := issue(t, s, "c1", "w1", apikey.Publishable, apikey.Live)
	if !strings.HasPrefix(pk.Secret, "lgr_live_pk_") {
		t.Fatalf("%q", pk.Secret)
	}
	var stored int64
	db.Raw(`SELECT COUNT(*) FROM "IdentityApiKeys" WHERE prefix = ? OR hash = ? OR name = ?`, k.Secret, k.Secret, k.Secret).Scan(&stored)
	var dump string
	db.Raw(`SELECT string_agg(t::text, ' ') FROM "IdentityApiKeys" t`).Scan(&dump)
	if stored != 0 || strings.Contains(dump, k.Secret) || strings.Contains(dump, pk.Secret) {
		t.Fatal("the key value is stored in the database; only its hash may be")
	}
	list, _ := s.List(kctx, "c1", "w1")
	for _, l := range list {
		if l.Secret != "" {
			t.Fatalf("listing exposed a secret: %+v", l)
		}
	}
}

func TestApiKey_AuthenticateReturnsTheKeysIdentityAndRejectsGarbage(t *testing.T) {
	s, _ := keySvc(t)
	k := issue(t, s, "c1", "w1", apikey.Secret, apikey.Test)
	c, err := s.Authenticate(kctx, k.Secret)
	if err != nil || c.ClientID != "c1" || c.WorkspaceID != "w1" || c.UserID != "usr-1" || c.Environment != "test" || c.Kind != "secret" || c.KeyID != k.ID {
		t.Fatalf("%+v %v", c, err)
	}
	for _, bad := range []string{"", "nonsense", "lgr_test_sk_short", "xxx_test_sk_" + strings.Repeat("a", 32), k.Secret + "x", strings.Replace(k.Secret, "lgr_test", "lgr_live", 1), "Bearer " + k.Secret} {
		if _, err := s.Authenticate(kctx, bad); !errors.Is(err, apikey.ErrInvalid) {
			t.Errorf("%q: want ErrInvalid, got %v", bad, err)
		}
	}
	// a different brand's service does not accept it
	other := &apikey.Service{DB: s.DB, Brand: "abc"}
	if _, err := other.Authenticate(kctx, k.Secret); !errors.Is(err, apikey.ErrInvalid) {
		t.Fatalf("%v", err)
	}
}

func TestApiKey_RevokedExpiredAndLiveLocked(t *testing.T) {
	s, _ := keySvc(t)
	now := time.Now()
	s.Now = func() time.Time { return now }

	k := issue(t, s, "c1", "w1", apikey.Secret, apikey.Test)
	if err := s.Revoke(kctx, "c1", k.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Authenticate(kctx, k.Secret); !errors.Is(err, apikey.ErrRevoked) {
		t.Fatalf("%v", err)
	}
	if err := s.Revoke(kctx, "c1", k.ID); !errors.Is(err, apikey.ErrNotFound) {
		t.Fatalf("a second revoke: %v", err)
	}

	short, _ := s.Issue(kctx, "u", "c1", "w1", "short", apikey.Secret, apikey.Test, nil, time.Hour)
	if _, err := s.Authenticate(kctx, short.Secret); err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Hour)
	if _, err := s.Authenticate(kctx, short.Secret); !errors.Is(err, apikey.ErrExpired) {
		t.Fatalf("%v", err)
	}

	allowed := false
	s.LiveAllowed = func(context.Context, string, string) bool { return allowed }
	if _, err := s.Issue(kctx, "u", "c1", "w1", "live", apikey.Secret, apikey.Live, nil, 0); !errors.Is(err, apikey.ErrLiveLocked) {
		t.Fatalf("live keys must be locked until go-live: %v", err)
	}
	if _, err := s.Issue(kctx, "u", "c1", "w1", "test", apikey.Secret, apikey.Test, nil, 0); err != nil {
		t.Fatalf("test keys are never locked: %v", err)
	}
	allowed = true
	live, err := s.Issue(kctx, "u", "c1", "w1", "live", apikey.Secret, apikey.Live, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	allowed = false // the workspace is suspended after go-live
	if _, err := s.Authenticate(kctx, live.Secret); !errors.Is(err, apikey.ErrLiveLocked) {
		t.Fatalf("an existing live key must stop working when live is withdrawn: %v", err)
	}
}

func TestApiKey_IssueValidatesItsInputs(t *testing.T) {
	s, _ := keySvc(t)
	for name, f := range map[string]func() error{
		"no workspace": func() error { _, err := s.Issue(kctx, "u", "c1", "", "n", "secret", "test", nil, 0); return err },
		"bad kind":     func() error { _, err := s.Issue(kctx, "u", "c1", "w", "n", "root", "test", nil, 0); return err },
		"bad env":      func() error { _, err := s.Issue(kctx, "u", "c1", "w", "n", "secret", "staging", nil, 0); return err },
		"no name":      func() error { _, err := s.Issue(kctx, "u", "c1", "w", " ", "secret", "test", nil, 0); return err },
	} {
		if f() == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestApiKey_ScopesAreCheckedWithWildcards(t *testing.T) {
	s, _ := keySvc(t)
	k, _ := s.Issue(kctx, "u", "c1", "w1", "limited", apikey.Secret, apikey.Test, []string{"wallets:*", "events:read"}, 0)
	c, _ := s.Authenticate(kctx, k.Secret)
	for scope, want := range map[string]bool{"wallets:read": true, "wallets:write": true, "events:read": true, "events:write": false, "keys:write": false} {
		if c.Allows(scope) != want {
			t.Errorf("%s: want %v", scope, want)
		}
	}
	all, _ := s.Issue(kctx, "u", "c1", "w1", "all", apikey.Secret, apikey.Test, nil, 0)
	ca, _ := s.Authenticate(kctx, all.Secret)
	if !ca.Allows("anything:at-all") {
		t.Error("the default scope is everything")
	}
}

func TestApiKey_AnotherClientCannotListRevokeOrRotate(t *testing.T) {
	s, _ := keySvc(t)
	k := issue(t, s, "c1", "w1", apikey.Secret, apikey.Test)
	if l, _ := s.List(kctx, "c2", ""); len(l) != 0 {
		t.Fatalf("c2 lists %d of c1's keys", len(l))
	}
	if err := s.Revoke(kctx, "c2", k.ID); !errors.Is(err, apikey.ErrNotFound) {
		t.Fatalf("%v", err)
	}
	if _, err := s.Rotate(kctx, "u", "c2", k.ID, 0); !errors.Is(err, apikey.ErrNotFound) {
		t.Fatalf("%v", err)
	}
	if _, err := s.Authenticate(kctx, k.Secret); err != nil {
		t.Fatalf("c1's key must still work: %v", err)
	}
}

func TestApiKey_RotationWithAndWithoutAGracePeriod(t *testing.T) {
	s, _ := keySvc(t)
	now := time.Now()
	s.Now = func() time.Time { return now }

	old := issue(t, s, "c1", "w1", apikey.Secret, apikey.Test)
	nk, err := s.Rotate(kctx, "usr-1", "c1", old.ID, time.Hour)
	if err != nil || nk.Secret == "" || nk.Secret == old.Secret || nk.Name != old.Name || nk.Environment != old.Environment {
		t.Fatalf("%+v %v", nk, err)
	}
	if _, err := s.Authenticate(kctx, old.Secret); err != nil {
		t.Fatalf("the old key must work during the grace period: %v", err)
	}
	if _, err := s.Authenticate(kctx, nk.Secret); err != nil {
		t.Fatalf("the new key: %v", err)
	}
	now = now.Add(2 * time.Hour)
	if _, err := s.Authenticate(kctx, old.Secret); !errors.Is(err, apikey.ErrExpired) {
		t.Fatalf("after the grace period: %v", err)
	}

	cut, _ := s.Rotate(kctx, "usr-1", "c1", nk.ID, 0)
	if _, err := s.Authenticate(kctx, nk.Secret); !errors.Is(err, apikey.ErrRevoked) {
		t.Fatalf("an immediate rotation revokes the old key: %v", err)
	}
	if _, err := s.Authenticate(kctx, cut.Secret); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Rotate(kctx, "u", "c1", nk.ID, 0); !errors.Is(err, apikey.ErrNotFound) {
		t.Fatalf("a revoked key cannot be rotated: %v", err)
	}
}

func TestApiKey_LastUsedIsRecordedButThrottled(t *testing.T) {
	s, db := keySvc(t)
	now := time.Now()
	s.Now = func() time.Time { return now }
	k := issue(t, s, "c1", "w1", apikey.Secret, apikey.Test)
	s.Authenticate(kctx, k.Secret)
	var first *time.Time
	db.Raw(`SELECT last_used_at FROM "IdentityApiKeys" WHERE id = ?`, k.ID).Scan(&first)
	if first == nil {
		t.Fatal("last_used_at not recorded")
	}
	now = now.Add(10 * time.Second)
	s.Authenticate(kctx, k.Secret)
	var second *time.Time
	db.Raw(`SELECT last_used_at FROM "IdentityApiKeys" WHERE id = ?`, k.ID).Scan(&second)
	if !second.Equal(*first) {
		t.Fatal("last_used_at was rewritten within the throttle window")
	}
	now = now.Add(2 * time.Minute)
	s.Authenticate(kctx, k.Secret)
	var third *time.Time
	db.Raw(`SELECT last_used_at FROM "IdentityApiKeys" WHERE id = ?`, k.ID).Scan(&third)
	if !third.After(*first) {
		t.Fatal("last_used_at did not advance after the window")
	}
}
