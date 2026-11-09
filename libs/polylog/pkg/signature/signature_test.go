package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"testing"
	"time"
)

func sign(secret, content string) []byte {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(content))
	return m.Sum(nil)
}

func TestAlchemyStyleHexBody(t *testing.T) {
	t.Setenv("TEST_SIGNING_KEY", "whsec")
	c, ok := FromMeta([]byte(`{"signature":{"header":"X-Alchemy-Signature","secretEnv":"TEST_SIGNING_KEY"}}`))
	if !ok {
		t.Fatal("no rule read")
	}
	body := []byte(`{"webhookId":"wh_1","event":{"activity":[{"value":1.5}]}}`)
	good := hex.EncodeToString(sign("whsec", string(body)))
	if err := Verify(c, body, map[string]string{"x-alchemy-signature": good}, time.Now()); err != nil {
		t.Fatalf("a good signature: %v", err)
	}
	if err := Verify(c, append(body, ' '), map[string]string{"X-Alchemy-Signature": good}, time.Now()); !errors.Is(err, ErrInvalid) {
		t.Fatalf("a changed body: %v", err)
	}
	if err := Verify(c, body, map[string]string{}, time.Now()); !errors.Is(err, ErrInvalid) {
		t.Fatalf("no signature: %v", err)
	}
	t.Setenv("TEST_SIGNING_KEY", "")
	if err := Verify(c, body, map[string]string{"X-Alchemy-Signature": good}, time.Now()); !errors.Is(err, ErrMisconfig) {
		t.Fatalf("no secret: %v", err)
	}
}

func TestPrefixBase64AndTimestamp(t *testing.T) {
	t.Setenv("TEST_SIGNING_KEY", "k")
	c := Config{Header: "Sig", SecretEnv: "TEST_SIGNING_KEY", Encoding: "base64", Prefix: "v1=", TimestampHeader: "Ts", ToleranceSeconds: 60}
	now := time.Unix(1_800_000_000, 0)
	ts := strconv.FormatInt(now.Unix(), 10)
	body := []byte("raw body")
	sig := "v1=" + base64.StdEncoding.EncodeToString(sign("k", ts+"."+string(body)))
	if err := Verify(c, body, map[string]string{"Sig": sig, "Ts": ts}, now); err != nil {
		t.Fatal(err)
	}
	if err := Verify(c, body, map[string]string{"Sig": sig, "Ts": ts}, now.Add(2*time.Minute)); !errors.Is(err, ErrStale) {
		t.Fatalf("a replay after the tolerance: %v", err)
	}
	if err := Verify(c, body, map[string]string{"Sig": sig[3:], "Ts": ts}, now); !errors.Is(err, ErrInvalid) {
		t.Fatalf("missing prefix: %v", err)
	}
}

func TestNoRule(t *testing.T) {
	if _, ok := FromMeta([]byte(`{"other":1}`)); ok {
		t.Fatal("a rule from meta without one")
	}
	if _, ok := FromMeta(nil); ok {
		t.Fatal("a rule from no meta")
	}
}
