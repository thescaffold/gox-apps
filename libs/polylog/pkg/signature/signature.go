// Package signature verifies a provider's webhook signature over the raw
// request body before polylog stores it (PLAN U6): Alchemy, a payment
// provider, envelope's per-vhost webhook. A source opts in through its meta:
//
//	{"signature": {
//	  "scheme": "hmac-sha256",            // or hmac-sha1, hmac-sha512 (default hmac-sha256)
//	  "header": "X-Alchemy-Signature",    // where the provider puts it
//	  "secretEnv": "ALCHEMY_SIGNING_KEY", // the env var holding the key (never the key itself)
//	  "encoding": "hex",                  // or base64 (default hex)
//	  "prefix": "",                       // e.g. "sha256=" (GitHub style), stripped before comparing
//	  "timestampHeader": "",              // when set, the signed content is "<timestamp>.<body>"
//	  "toleranceSeconds": 300             // and the timestamp (unix seconds) must be this fresh
//	}}
//
// The secret lives in the environment, so a database row never holds it.
package signature

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is a source's signature rule.
type Config struct {
	Scheme           string `json:"scheme"`
	Header           string `json:"header"`
	SecretEnv        string `json:"secretEnv"`
	Encoding         string `json:"encoding"`
	Prefix           string `json:"prefix"`
	TimestampHeader  string `json:"timestampHeader"`
	ToleranceSeconds int    `json:"toleranceSeconds"`
}

// Errors. A caller answers ErrInvalid with 401 and stores nothing.
var (
	ErrInvalid    = errors.New("signature: missing or invalid")
	ErrStale      = errors.New("signature: timestamp outside the tolerance")
	ErrMisconfig  = errors.New("signature: the source's rule is incomplete (header and secretEnv, with the env var set)")
	ErrNoRule     = errors.New("signature: this source has no signature rule")
	errBadSchemes = errors.New("signature: unknown scheme")
)

// FromMeta reads the rule from a source's meta JSON; ok is false when there is none.
func FromMeta(meta json.RawMessage) (Config, bool) {
	if len(meta) == 0 {
		return Config{}, false
	}
	var m struct {
		Signature *Config `json:"signature"`
	}
	if json.Unmarshal(meta, &m) != nil || m.Signature == nil {
		return Config{}, false
	}
	return *m.Signature, true
}

func newHash(scheme string) (func() hash.Hash, error) {
	switch strings.ToLower(scheme) {
	case "", "hmac-sha256":
		return sha256.New, nil
	case "hmac-sha1":
		return sha1.New, nil
	case "hmac-sha512":
		return sha512.New, nil
	}
	return nil, errBadSchemes
}

func header(headers map[string]string, name string) string {
	for k, v := range headers {
		if strings.EqualFold(k, name) {
			return v
		}
	}
	return ""
}

// Verify checks body against the signature in headers under c, at now.
func Verify(c Config, body []byte, headers map[string]string, now time.Time) error {
	h, err := newHash(c.Scheme)
	if err != nil || c.Header == "" || c.SecretEnv == "" {
		return ErrMisconfig
	}
	secret := os.Getenv(c.SecretEnv)
	if secret == "" {
		return ErrMisconfig
	}
	got := strings.TrimSpace(header(headers, c.Header))
	if c.Prefix != "" {
		if !strings.HasPrefix(got, c.Prefix) {
			return ErrInvalid
		}
		got = strings.TrimPrefix(got, c.Prefix)
	}
	if got == "" {
		return ErrInvalid
	}
	var sent []byte
	if strings.EqualFold(c.Encoding, "base64") {
		sent, err = base64.StdEncoding.DecodeString(got)
	} else {
		sent, err = hex.DecodeString(strings.ToLower(got))
	}
	if err != nil {
		return ErrInvalid
	}
	mac := hmac.New(h, []byte(secret))
	if c.TimestampHeader != "" {
		ts := strings.TrimSpace(header(headers, c.TimestampHeader))
		sec, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			return ErrInvalid
		}
		tol := c.ToleranceSeconds
		if tol <= 0 {
			tol = 300
		}
		if math.Abs(float64(now.Unix()-sec)) > float64(tol) {
			return ErrStale
		}
		mac.Write([]byte(ts + "."))
	}
	mac.Write(body)
	if !hmac.Equal(mac.Sum(nil), sent) {
		return ErrInvalid
	}
	return nil
}
