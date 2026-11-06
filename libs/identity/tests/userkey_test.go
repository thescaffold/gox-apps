package tests

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"

	appuserkey "github.com/thescaffold/gox-apps/libs/identity/app/userkey"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/userkey"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

func edKey(t *testing.T, comment string) string {
	t.Helper()
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	sp, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sp))) + " " + comment
}

func rsaKey(t *testing.T, bits int) string {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		t.Fatal(err)
	}
	sp, err := ssh.NewPublicKey(&k.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return string(ssh.MarshalAuthorizedKey(sp))
}

func gpgKey(t *testing.T, private bool) string {
	t.Helper()
	e, err := openpgp.NewEntity("Ada Lovelace", "", "Ada@Example.test", nil)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	block := openpgp.PublicKeyType
	if private {
		block = openpgp.PrivateKeyType
	}
	w, err := armor.Encode(&buf, block, nil)
	if err != nil {
		t.Fatal(err)
	}
	if private {
		err = e.SerializePrivate(w, nil)
	} else {
		err = e.Serialize(w)
	}
	if err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
	return buf.String()
}

func TestUserKeys_ParseAcceptsPublicKeysOnly(t *testing.T) {
	now := time.Now()
	ed := edKey(t, "ada@laptop")
	p, err := userkey.Parse(ed, now)
	if err != nil || p.Format != userkey.FormatSSH || p.Algorithm != "ssh-ed25519" || !strings.HasPrefix(p.Fingerprint, "SHA256:") || p.Comment != "ada@laptop" {
		t.Fatalf("ed25519: %+v %v", p, err)
	}
	if strings.Contains(p.PublicKey, "ada@laptop") {
		t.Fatalf("the stored key is normalized without its comment: %q", p.PublicKey)
	}
	if p, err := userkey.Parse(rsaKey(t, 2048), now); err != nil || p.Bits != 2048 {
		t.Fatalf("rsa 2048: %+v %v", p, err)
	}
	for name, c := range map[string]struct {
		text string
		want error
	}{
		"rsa 1024":       {rsaKey(t, 1024), userkey.ErrWeak},
		"openssh secret": {"-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAA\n-----END OPENSSH PRIVATE KEY-----", userkey.ErrPrivate},
		"gpg secret":     {gpgKey(t, true), userkey.ErrPrivate},
		"garbage":        {"ssh-ed25519 not-base64!!", userkey.ErrInvalid},
		"two lines":      {ed + "\n" + edKey(t, "other"), userkey.ErrInvalid},
		"empty":          {"   ", userkey.ErrInvalid},
	} {
		if _, err := userkey.Parse(c.text, now); !errors.Is(err, c.want) {
			t.Errorf("%s: got %v, want %v", name, err, c.want)
		}
	}
	// Parsed with our clock a little behind the key's creation (its maker's clock runs ahead).
	g, err := userkey.Parse(gpgKey(t, false), now.Add(-30*time.Second))
	if err != nil || g.Format != userkey.FormatGPG || len(g.Fingerprint) != 40 || len(g.KeyID) != 16 || len(g.Emails) != 1 || g.Emails[0] != "ada@example.test" || g.Comment != "Ada Lovelace" {
		t.Fatalf("gpg: %+v %v", g, err)
	}
	if _, err := userkey.Parse(gpgKey(t, false), now.Add(-2*time.Hour)); !errors.Is(err, userkey.ErrGPGUnusable) {
		t.Fatalf("a key from hours in the future: %v", err)
	}
}

func TestUserKeys_StoreRules(t *testing.T) {
	db := identityTestDB(t)
	svc := &userkey.Service{DB: db}
	ctx := context.Background()
	ada, bob := "u-ada-"+uuid.NewString()[:8], "u-bob-"+uuid.NewString()[:8]

	var mu sync.Mutex
	var seen []string
	userkey.SetObserver(func(_ context.Context, e userkey.Event) {
		mu.Lock()
		seen = append(seen, e.Action+":"+e.Key.Usage)
		mu.Unlock()
	})
	defer userkey.SetObserver(nil)

	ed := edKey(t, "ada@laptop")
	auth, err := svc.Add(ctx, ada, userkey.UsageAuth, "", ed)
	if err != nil || auth.Title != "ada@laptop" {
		t.Fatalf("add auth: %+v %v", auth, err)
	}
	// The same key signs for the same person too.
	if _, err := svc.Add(ctx, ada, userkey.UsageSigning, "Laptop signing", ed); err != nil {
		t.Fatalf("add the same key for signing: %v", err)
	}
	if _, err := svc.Add(ctx, ada, userkey.UsageAuth, "again", ed); !errors.Is(err, userkey.ErrDuplicate) {
		t.Fatalf("duplicate: %v", err)
	}
	// Nobody else can claim it, for either use.
	if _, err := svc.Add(ctx, bob, userkey.UsageAuth, "", ed); !errors.Is(err, userkey.ErrTaken) {
		t.Fatalf("another person's auth key: %v", err)
	}
	if _, err := svc.Add(ctx, bob, userkey.UsageSigning, "", ed); !errors.Is(err, userkey.ErrTaken) {
		t.Fatalf("another person's signing key: %v", err)
	}
	// GPG keys sign only.
	gpg := gpgKey(t, false)
	if _, err := svc.Add(ctx, ada, userkey.UsageAuth, "", gpg); !errors.Is(err, userkey.ErrGPGAuth) {
		t.Fatalf("gpg for auth: %v", err)
	}
	gk, err := svc.Add(ctx, ada, userkey.UsageSigning, "", gpg)
	if err != nil || gk.KeyID == "" || len(gk.Emails) != 1 {
		t.Fatalf("gpg signing: %+v %v", gk, err)
	}
	if _, err := svc.Add(ctx, ada, "deploy", "", edKey(t, "x")); !errors.Is(err, userkey.ErrInvalid) {
		t.Fatalf("unknown usage: %v", err)
	}

	// The git SSH server's question.
	found, err := svc.LookupAuth(ctx, auth.Fingerprint)
	if err != nil || found.UserID != ada {
		t.Fatalf("lookup: %+v %v", found, err)
	}
	if _, err := svc.LookupAuth(ctx, "SHA256:nobody"); !errors.Is(err, userkey.ErrNotFound) {
		t.Fatalf("lookup unknown: %v", err)
	}
	if err := svc.Used(ctx, auth.ID); err != nil {
		t.Fatal(err)
	}

	list, _ := svc.List(ctx, ada)
	if len(list) != 3 {
		t.Fatalf("ada has three keys: %d", len(list))
	}
	signing, _ := svc.SigningKeys(ctx, ada)
	if len(signing) != 2 {
		t.Fatalf("two signing keys: %d", len(signing))
	}

	// One's own keys only, through the routes too.
	c := appuserkey.NewControllerForTest(svc)
	if out := c.Remove(&appuserkey.IDReq{Ctx: ntxctx.NTXContext{UserID: bob}, Id: auth.ID}); out.Code() != 404 {
		t.Fatalf("bob removes ada's key: %d", out.Code())
	}
	if out := c.List(&appuserkey.ListReq{Ctx: ntxctx.NTXContext{UserID: bob}}); out.Code() != 200 || strings.Contains(string(mustJSON(out.Data())), auth.ID) {
		t.Fatalf("bob lists ada's key: %d", out.Code())
	}
	if out := c.Add(&appuserkey.AddReq{Ctx: ntxctx.NTXContext{}, Usage: "auth", Key: edKey(t, "x")}); out.Code() != 401 {
		t.Fatalf("signed out: %d", out.Code())
	}
	if out := c.Add(&appuserkey.AddReq{Ctx: ntxctx.NTXContext{UserID: bob}, Usage: "auth", Key: gpgKey(t, true)}); out.Code() != 400 {
		t.Fatalf("private key: %d", out.Code())
	}
	if out := c.Remove(&appuserkey.IDReq{Ctx: ntxctx.NTXContext{UserID: ada}, Id: auth.ID}); out.Code() != 200 {
		t.Fatalf("ada removes her key: %d", out.Code())
	}
	if _, err := svc.LookupAuth(ctx, auth.Fingerprint); !errors.Is(err, userkey.ErrNotFound) {
		t.Fatalf("a removed key still authenticates: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if strings.Join(seen, ",") != "key.add:auth,key.add:signing,key.add:signing,key.remove:auth" {
		t.Fatalf("events: %v", seen)
	}
}
