package tests

import (
	"testing"

	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

func (r *rig) patchAttrs(userID, attrType string, body map[string]any) int {
	r.t.Helper()
	rep := decode(r.t, r.ctl.UpdateAttribute(&auth.UpdateAttributesDto{Ctx: ntxctx.NTXContext{UserID: userID}, Type: attrType, Body: body}))
	return rep.code
}

func (r *rig) attr(userID, key string) (string, bool) {
	var v string
	res := r.db.Raw(`SELECT value FROM "IdentityAttributes" WHERE user_id = ? AND key = ?`, userID, key).Scan(&v)
	return v, res.RowsAffected > 0
}

func TestUserCannotWriteServerOwnedAttributes(t *testing.T) {
	r := newRig(t)
	ada, _ := r.person("ada")

	if code := r.patchAttrs(ada.Id, "preference", map[string]any{"theme": "dark"}); code != 200 {
		t.Fatalf("a preference must be writable, got %d", code)
	}
	if v, ok := r.attr(ada.Id, "theme"); !ok || v != "dark" {
		t.Fatalf("theme not stored: %q %v", v, ok)
	}

	for name, tc := range map[string]struct {
		typ  string
		body map[string]any
	}{
		"closed_at":       {"user", map[string]any{"closed_at": "x"}},
		"suspended_at":    {"user", map[string]any{"suspended_at": "x"}},
		"referrer_id":     {"user", map[string]any{"referrer_id": "someone"}},
		"underscore key":  {"user", map[string]any{"_pin_hash": "x"}},
		"reserved type":   {"security", map[string]any{"anything": "x"}},
		"underscore type": {"_internal", map[string]any{"anything": "x"}},
	} {
		if code := r.patchAttrs(ada.Id, tc.typ, tc.body); code != 403 {
			t.Errorf("%s: want 403, got %d", name, code)
		}
	}
	for _, k := range []string{"closed_at", "suspended_at", "referrer_id", "_pin_hash", "anything"} {
		if _, ok := r.attr(ada.Id, k); ok {
			t.Errorf("%s was written despite the rejection", k)
		}
	}
}

func TestOneReservedKeyRejectsTheWholeRequestAndNothingIsWritten(t *testing.T) {
	r := newRig(t)
	ada, _ := r.person("ada")
	if code := r.patchAttrs(ada.Id, "user", map[string]any{"nickname": "ad", "closed_at": "x"}); code != 403 {
		t.Fatalf("got %d", code)
	}
	if _, ok := r.attr(ada.Id, "nickname"); ok {
		t.Fatal("a rejected request must change nothing")
	}
}

func TestHostsCanReserveMoreKeys(t *testing.T) {
	r := newRig(t)
	ada, _ := r.person("ada")
	if code := r.patchAttrs(ada.Id, "user", map[string]any{"pin_hash": "x"}); code != 200 {
		t.Fatalf("pin_hash is not reserved until a host reserves it, got %d", code)
	}
	auth.ReserveAttributeKeys("pin_hash2")
	if code := r.patchAttrs(ada.Id, "user", map[string]any{"pin_hash2": "x"}); code != 403 {
		t.Fatalf("got %d", code)
	}
}
