// Package userkey is a person's public keys (U-S22): SSH keys that
// authenticate git over SSH, and SSH or GPG keys that sign commits.
//
// Only public keys are accepted and kept; a pasted private key is refused
// without being stored. Weak keys are refused (DSA, RSA under 2048 bits), as
// are revoked or expired GPG keys and GPG keys that cannot sign. For each use
// a key belongs to one person across the service: the git SSH server
// identifies the person by an authentication key (LookupAuth), and a signature
// must verify as one person's. One person may use the same SSH key for both.
package userkey

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

// Usages and formats.
const (
	UsageAuth    = "auth"
	UsageSigning = "signing"
	FormatSSH    = "ssh"
	FormatGPG    = "gpg"
)

// MaxKeys is how many keys one person may hold.
const MaxKeys = 100

// Errors a caller can match. Each reads as a sentence a person can act on.
var (
	ErrInvalid     = errors.New("that is not a public key we can use")
	ErrPrivate     = errors.New("that is a private key; paste the public key (the .pub file, or gpg --armor --export) instead, and keep the private key to yourself")
	ErrWeak        = errors.New("that key is too weak; use Ed25519, ECDSA, or RSA of at least 2048 bits")
	ErrGPGAuth     = errors.New("GPG keys sign commits only; use an SSH key for git over SSH")
	ErrGPGUnusable = errors.New("that GPG key is revoked, expired or cannot sign")
	ErrTaken       = errors.New("this key is already in use by an account")
	ErrDuplicate   = errors.New("you have already added this key")
	ErrTooMany     = errors.New("you have reached the limit of 100 keys; remove one first")
	ErrNotFound    = errors.New("key not found")
)

// Key is one public key.
type Key struct {
	ID          string          `gorm:"column:id;primaryKey" json:"id"`
	UserID      string          `gorm:"column:user_id" json:"-"`
	Usage       string          `gorm:"column:usage" json:"usage"`
	Format      string          `gorm:"column:format" json:"format"`
	Title       string          `gorm:"column:title" json:"title"`
	PublicKey   string          `gorm:"column:public_key" json:"publicKey"`
	Fingerprint string          `gorm:"column:fingerprint" json:"fingerprint"`
	Algorithm   string          `gorm:"column:algorithm" json:"algorithm"`
	Bits        int             `gorm:"column:bits" json:"bits,omitempty"`
	KeyID       string          `gorm:"column:key_id" json:"keyId,omitempty"`
	EmailsJSON  json.RawMessage `gorm:"column:emails;type:jsonb" json:"-"`
	Emails      []string        `gorm:"-" json:"emails,omitempty"`
	ExpiresAt   *time.Time      `gorm:"column:expires_at" json:"expiresAt,omitempty"`
	LastUsedAt  *time.Time      `gorm:"column:last_used_at" json:"lastUsedAt,omitempty"`
	CreatedAt   time.Time       `gorm:"column:created_at" json:"createdAt"`
}

func (Key) TableName() string { return "IdentityUserKeys" }

func (k *Key) decode() {
	if len(k.EmailsJSON) > 0 {
		_ = json.Unmarshal(k.EmailsJSON, &k.Emails)
	}
}

// Parsed is what a pasted key turned out to be.
type Parsed struct {
	Format      string
	PublicKey   string // normalized: "type base64" for SSH, the armored block for GPG
	Fingerprint string // SHA256:… for SSH, 40 hex digits for GPG
	Algorithm   string
	Bits        int
	KeyID       string
	Emails      []string
	ExpiresAt   *time.Time
	Comment     string // the SSH key's trailing comment, or the GPG key's first user ID
}

var sshAllowed = map[string]string{
	ssh.KeyAlgoED25519:    "ssh-ed25519",
	ssh.KeyAlgoSKED25519:  "sk-ssh-ed25519",
	ssh.KeyAlgoECDSA256:   "ecdsa-sha2-nistp256",
	ssh.KeyAlgoECDSA384:   "ecdsa-sha2-nistp384",
	ssh.KeyAlgoECDSA521:   "ecdsa-sha2-nistp521",
	ssh.KeyAlgoSKECDSA256: "sk-ecdsa-sha2-nistp256",
	ssh.KeyAlgoRSA:        "ssh-rsa",
}

// Parse reads a pasted public key: an SSH authorized_keys line or an armored
// GPG public key block.
func Parse(text string, now time.Time) (Parsed, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Parsed{}, ErrInvalid
	}
	if strings.Contains(text, "PRIVATE KEY") {
		return Parsed{}, ErrPrivate
	}
	if strings.HasPrefix(text, "-----BEGIN PGP PUBLIC KEY BLOCK-----") {
		return parseGPG(text, now)
	}
	return parseSSH(text)
}

func parseSSH(text string) (Parsed, error) {
	if strings.ContainsAny(strings.TrimSpace(text), "\r\n") {
		return Parsed{}, fmt.Errorf("%w: paste one key, on one line", ErrInvalid)
	}
	pub, comment, _, rest, err := ssh.ParseAuthorizedKey([]byte(text))
	if err != nil || len(bytes.TrimSpace(rest)) > 0 {
		return Parsed{}, ErrInvalid
	}
	if _, isCert := pub.(*ssh.Certificate); isCert {
		return Parsed{}, fmt.Errorf("%w: certificates are not accepted; paste the key itself", ErrInvalid)
	}
	t := pub.Type()
	if t == ssh.KeyAlgoDSA {
		return Parsed{}, ErrWeak
	}
	alg, ok := sshAllowed[t]
	if !ok {
		return Parsed{}, ErrInvalid
	}
	bits := 0
	switch t {
	case ssh.KeyAlgoRSA:
		cpk, ok := pub.(ssh.CryptoPublicKey)
		if !ok {
			return Parsed{}, ErrInvalid
		}
		r, ok := cpk.CryptoPublicKey().(*rsa.PublicKey)
		if !ok {
			return Parsed{}, ErrInvalid
		}
		bits = r.N.BitLen()
		if bits < 2048 {
			return Parsed{}, ErrWeak
		}
	case ssh.KeyAlgoED25519, ssh.KeyAlgoSKED25519:
		bits = 256
	case ssh.KeyAlgoECDSA256, ssh.KeyAlgoSKECDSA256:
		bits = 256
	case ssh.KeyAlgoECDSA384:
		bits = 384
	case ssh.KeyAlgoECDSA521:
		bits = 521
	}
	return Parsed{
		Format:      FormatSSH,
		PublicKey:   strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pub))),
		Fingerprint: ssh.FingerprintSHA256(pub),
		Algorithm:   alg,
		Bits:        bits,
		Comment:     strings.TrimSpace(comment),
	}, nil
}

var gpgAlgos = map[packet.PublicKeyAlgorithm]string{
	packet.PubKeyAlgoRSA:         "RSA",
	packet.PubKeyAlgoRSASignOnly: "RSA",
	packet.PubKeyAlgoDSA:         "DSA",
	packet.PubKeyAlgoECDSA:       "ECDSA",
	packet.PubKeyAlgoEdDSA:       "EdDSA",
	packet.PubKeyAlgoEd25519:     "Ed25519",
	packet.PubKeyAlgoEd448:       "Ed448",
}

func parseGPG(text string, now time.Time) (Parsed, error) {
	ents, err := openpgp.ReadArmoredKeyRing(strings.NewReader(text))
	if err != nil || len(ents) == 0 {
		return Parsed{}, ErrInvalid
	}
	if len(ents) > 1 {
		return Parsed{}, fmt.Errorf("%w: paste one key at a time", ErrInvalid)
	}
	e := ents[0]
	if e.PrivateKey != nil {
		return Parsed{}, ErrPrivate
	}
	pk := e.PrimaryKey
	// A key made moments ago on a computer whose clock runs ahead is not yet
	// valid by ours: allow an hour of skew on its creation time.
	at := now
	if pk.CreationTime.After(now) {
		if pk.CreationTime.Sub(now) > time.Hour {
			return Parsed{}, ErrGPGUnusable
		}
		at = pk.CreationTime
	}
	if e.Revoked(at) {
		return Parsed{}, ErrGPGUnusable
	}
	if _, ok := e.SigningKey(at); !ok {
		return Parsed{}, ErrGPGUnusable
	}
	alg, ok := gpgAlgos[pk.PubKeyAlgo]
	if !ok {
		return Parsed{}, ErrInvalid
	}
	bits := 0
	if b, err := pk.BitLength(); err == nil {
		bits = int(b)
	}
	if (alg == "RSA" && bits < 2048) || alg == "DSA" {
		return Parsed{}, ErrWeak
	}
	var emails []string
	comment := ""
	seen := map[string]bool{}
	for _, id := range e.Identities {
		if id.UserId == nil {
			continue
		}
		if comment == "" {
			comment = id.UserId.Name
		}
		if m := strings.ToLower(strings.TrimSpace(id.UserId.Email)); m != "" && !seen[m] {
			seen[m] = true
			emails = append(emails, m)
		}
	}
	if p := e.PrimaryIdentity(); p != nil && p.UserId != nil && p.UserId.Name != "" {
		comment = p.UserId.Name
	}
	var expires *time.Time
	if p := e.PrimaryIdentity(); p != nil && p.SelfSignature != nil && p.SelfSignature.KeyLifetimeSecs != nil && *p.SelfSignature.KeyLifetimeSecs > 0 {
		t := pk.CreationTime.Add(time.Duration(*p.SelfSignature.KeyLifetimeSecs) * time.Second).UTC()
		expires = &t
	}
	return Parsed{
		Format:      FormatGPG,
		PublicKey:   text,
		Fingerprint: strings.ToUpper(hex.EncodeToString(pk.Fingerprint)),
		Algorithm:   alg,
		Bits:        bits,
		KeyID:       strings.ToUpper(pk.KeyIdString()),
		Emails:      emails,
		ExpiresAt:   expires,
		Comment:     comment,
	}, nil
}

// Service is the store. Now is a seam for tests.
type Service struct {
	DB  *gorm.DB
	Now func() time.Time
}

func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

// Add stores a key for userID. usage is auth or signing; title defaults to
// the key's own comment.
func (s *Service) Add(ctx context.Context, userID, usage, title, text string) (Key, error) {
	if usage != UsageAuth && usage != UsageSigning {
		return Key{}, fmt.Errorf("%w: choose authentication or signing", ErrInvalid)
	}
	p, err := Parse(text, s.now())
	if err != nil {
		return Key{}, err
	}
	if p.Format == FormatGPG && usage == UsageAuth {
		return Key{}, ErrGPGAuth
	}
	title = strings.TrimSpace(title)
	if title == "" {
		title = p.Comment
	}
	if title == "" {
		title = p.Algorithm + " key"
	}
	if len(title) > 120 {
		title = title[:120]
	}
	var k Key
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// One person's keys are counted and checked under their own lock.
		if err := tx.Exec(`SELECT pg_advisory_xact_lock(hashtext(?))`, "identity-user-keys:"+userID).Error; err != nil {
			return err
		}
		var n int64
		if err := tx.Model(&Key{}).Where(`"user_id" = ?`, userID).Count(&n).Error; err != nil {
			return err
		}
		if n >= MaxKeys {
			return ErrTooMany
		}
		var dup []Key
		if err := tx.Where(`"fingerprint" = ? AND "usage" = ?`, p.Fingerprint, usage).Limit(5).Find(&dup).Error; err != nil {
			return err
		}
		for _, d := range dup {
			if d.UserID == userID {
				return ErrDuplicate
			}
			return ErrTaken
		}
		emails, _ := json.Marshal(p.Emails)
		k = Key{
			ID: uuid.NewString(), UserID: userID, Usage: usage, Format: p.Format, Title: title,
			PublicKey: p.PublicKey, Fingerprint: p.Fingerprint, Algorithm: p.Algorithm, Bits: p.Bits,
			KeyID: p.KeyID, EmailsJSON: emails, Emails: p.Emails, ExpiresAt: p.ExpiresAt, CreatedAt: s.now(),
		}
		return tx.Create(&k).Error
	})
	if err != nil {
		return Key{}, err
	}
	Emit(ctx, Event{Action: "key.add", UserID: userID, Key: k})
	return k, nil
}

// List returns the person's keys, newest first.
func (s *Service) List(ctx context.Context, userID string) ([]Key, error) {
	var ks []Key
	err := s.DB.WithContext(ctx).Where(`"user_id" = ?`, userID).Order(`"created_at" DESC`).Find(&ks).Error
	for i := range ks {
		ks[i].decode()
	}
	return ks, err
}

// Remove deletes one of the person's own keys.
func (s *Service) Remove(ctx context.Context, userID, id string) error {
	var k Key
	err := s.DB.WithContext(ctx).Where(`"id" = ? AND "user_id" = ?`, id, userID).Take(&k).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if err := s.DB.WithContext(ctx).Where(`"id" = ? AND "user_id" = ?`, id, userID).Delete(&Key{}).Error; err != nil {
		return err
	}
	k.decode()
	Emit(ctx, Event{Action: "key.remove", UserID: userID, Key: k})
	return nil
}

// LookupAuth finds the authentication key with this SSH fingerprint (the git
// SSH server's question: who is connecting?).
func (s *Service) LookupAuth(ctx context.Context, fingerprint string) (Key, error) {
	var ks []Key
	err := s.DB.WithContext(ctx).Where(`"fingerprint" = ? AND "usage" = ?`, fingerprint, UsageAuth).Limit(1).Find(&ks).Error
	if err != nil {
		return Key{}, err
	}
	if len(ks) == 0 {
		return Key{}, ErrNotFound
	}
	return ks[0], nil
}

// Used records that the key was just used.
func (s *Service) Used(ctx context.Context, id string) error {
	return s.DB.WithContext(ctx).Model(&Key{}).Where(`"id" = ?`, id).Update("last_used_at", s.now()).Error
}

// SigningKeys returns the person's signing keys (commit verification).
func (s *Service) SigningKeys(ctx context.Context, userID string) ([]Key, error) {
	var ks []Key
	err := s.DB.WithContext(ctx).Where(`"user_id" = ? AND "usage" = ?`, userID, UsageSigning).Find(&ks).Error
	for i := range ks {
		ks[i].decode()
	}
	return ks, err
}

// Event is a change a host may record (an audit log).
type Event struct {
	Action string // key.add, key.remove
	UserID string
	Key    Key
}

var (
	mu       sync.RWMutex
	observer func(context.Context, Event)
)

// SetObserver registers the host's recorder (nil turns it off).
func SetObserver(f func(context.Context, Event)) {
	mu.Lock()
	defer mu.Unlock()
	observer = f
}

// Emit hands e to the observer, if any.
func Emit(ctx context.Context, e Event) {
	mu.RLock()
	f := observer
	mu.RUnlock()
	if f != nil {
		f(ctx, e)
	}
}
