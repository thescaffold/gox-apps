package tests

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/awesome-goose/goose/types"
	apikeyapp "github.com/thescaffold/gox-apps/libs/identity/app/apikey"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

func kbody(o types.Output) (int, string) {
	b, _ := json.Marshal(o.Data())
	return o.Code(), string(b)
}

func TestApiKeyController_IssueListRotateRevoke(t *testing.T) {
	s, _ := keySvc(t)
	c := apikeyapp.NewControllerForTest(s)
	ctx := ntxctx.NTXContext{UserID: "usr-1", ClientID: "c1", WorkspaceID: "w1"}
	other := ntxctx.NTXContext{UserID: "usr-2", ClientID: "c2", WorkspaceID: "w2"}

	if code, _ := kbody(c.Issue(&apikeyapp.IssueReq{Name: "x", Ctx: ntxctx.NTXContext{}})); code != 401 {
		t.Fatalf("no scope: %d", code)
	}
	if code, _ := kbody(c.Issue(&apikeyapp.IssueReq{Name: "x", Kind: "root", Ctx: ctx})); code != 400 {
		t.Fatalf("bad kind: %d", code)
	}
	code, b := kbody(c.Issue(&apikeyapp.IssueReq{Name: "backend", Ctx: ctx}))
	if code != 201 || !strings.Contains(b, `"secret":"lgr_test_sk_`) {
		t.Fatalf("defaults are a secret test key: %d %s", code, b)
	}
	var created struct{ Data struct{ ID, Secret string } }
	_ = json.Unmarshal([]byte(b), &created)

	if _, b := kbody(c.List(&apikeyapp.ListReq{Ctx: ctx})); !strings.Contains(b, "backend") || strings.Contains(b, created.Data.Secret) {
		t.Fatalf("list: %s", b)
	}
	if _, b := kbody(c.List(&apikeyapp.ListReq{Ctx: other})); strings.Contains(b, "backend") {
		t.Fatalf("another client sees the key: %s", b)
	}
	if code, _ := kbody(c.Revoke(&apikeyapp.IDReq{Id: created.Data.ID, Ctx: other})); code != 404 {
		t.Fatalf("another client revoked it: %d", code)
	}
	code, b = kbody(c.Rotate(&apikeyapp.IDReq{Id: created.Data.ID, Grace: "60", Ctx: ctx}))
	if code != 201 || !strings.Contains(b, `"secret":"lgr_test_sk_`) {
		t.Fatalf("rotate: %d %s", code, b)
	}
	if code, _ := kbody(c.Revoke(&apikeyapp.IDReq{Id: created.Data.ID, Ctx: ctx})); code != 200 {
		t.Fatalf("revoke: %d", code)
	}
	if code, _ := kbody(c.Revoke(&apikeyapp.IDReq{Id: created.Data.ID, Ctx: ctx})); code != 404 {
		t.Fatalf("revoke twice: %d", code)
	}
}
