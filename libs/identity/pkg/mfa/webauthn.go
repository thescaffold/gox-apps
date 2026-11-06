package mfa

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

// Keys is the WebAuthn relying party (security keys).
type Keys struct{ wa *webauthn.WebAuthn }

// KeysConfig names the relying party: the domain keys are bound to (RPID)
// and the origins a browser may use them from.
type KeysConfig struct {
	RPID        string
	DisplayName string
	Origins     []string
}

// NewKeys builds the relying party.
func NewKeys(c KeysConfig) (*Keys, error) {
	wa, err := webauthn.New(&webauthn.Config{RPID: c.RPID, RPDisplayName: c.DisplayName, RPOrigins: c.Origins})
	if err != nil {
		return nil, err
	}
	return &Keys{wa: wa}, nil
}

// KeysFromEnv reads WEBAUTHN_RP_ID and WEBAUTHN_ORIGINS (comma-separated),
// defaulting both from APP_URL and BASE_URL; the display name is
// BUSINESS_NAME or APP_NAME. It returns nil when no origin is known.
func KeysFromEnv() (*Keys, error) {
	var origins []string
	if v := strings.TrimSpace(os.Getenv("WEBAUTHN_ORIGINS")); v != "" {
		for _, o := range strings.Split(v, ",") {
			if o = strings.TrimRight(strings.TrimSpace(o), "/"); o != "" {
				origins = append(origins, o)
			}
		}
	} else {
		seen := map[string]bool{}
		for _, k := range []string{"APP_URL", "BASE_URL"} {
			u, err := url.Parse(strings.TrimSpace(os.Getenv(k)))
			if err != nil || u.Scheme == "" || u.Host == "" {
				continue
			}
			o := u.Scheme + "://" + u.Host
			if !seen[o] {
				seen[o] = true
				origins = append(origins, o)
			}
		}
	}
	if len(origins) == 0 {
		return nil, nil
	}
	rpID := strings.TrimSpace(os.Getenv("WEBAUTHN_RP_ID"))
	if rpID == "" {
		u, _ := url.Parse(origins[0])
		rpID = u.Hostname()
	}
	name := os.Getenv("BUSINESS_NAME")
	if name == "" {
		name = os.Getenv("APP_NAME")
	}
	if name == "" {
		name = rpID
	}
	return NewKeys(KeysConfig{RPID: rpID, DisplayName: name, Origins: origins})
}

// keyUser adapts a person and their keys to the library.
type keyUser struct {
	id, name, display string
	creds             []webauthn.Credential
}

func (u keyUser) WebAuthnID() []byte                         { return []byte(u.id) }
func (u keyUser) WebAuthnName() string                       { return u.name }
func (u keyUser) WebAuthnDisplayName() string                { return u.display }
func (u keyUser) WebAuthnCredentials() []webauthn.Credential { return u.creds }

func (s *Service) keyUser(ctx context.Context, userID, name, display string) (keyUser, []Factor, error) {
	var fs []Factor
	err := s.DB.WithContext(ctx).Where(`"user_id" = ? AND "kind" = ? AND "confirmed_at" IS NOT NULL`, userID, KindWebAuthn).Find(&fs).Error
	if err != nil {
		return keyUser{}, nil, err
	}
	u := keyUser{id: userID, name: name, display: display}
	for _, f := range fs {
		var c webauthn.Credential
		if json.Unmarshal(f.Credential, &c) == nil {
			u.creds = append(u.creds, c)
		}
	}
	if u.name == "" {
		u.name = userID
	}
	if u.display == "" {
		u.display = u.name
	}
	return u, fs, nil
}

// BeginKey starts adding a security key: it returns the options for the
// browser's navigator.credentials.create and the token FinishKey needs.
// Keys the person already has are excluded.
func (s *Service) BeginKey(ctx context.Context, userID, name, display string) (json.RawMessage, string, error) {
	if s.Keys == nil {
		return nil, "", ErrKeysUnavailable
	}
	u, _, err := s.keyUser(ctx, userID, name, display)
	if err != nil {
		return nil, "", err
	}
	var exclude []protocol.CredentialDescriptor
	for _, c := range u.creds {
		exclude = append(exclude, c.Descriptor())
	}
	creation, session, err := s.Keys.wa.BeginRegistration(u,
		webauthn.WithExclusions(exclude),
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			ResidentKey:      protocol.ResidentKeyRequirementDiscouraged,
			UserVerification: protocol.VerificationDiscouraged,
		}))
	if err != nil {
		return nil, "", err
	}
	raw, err := json.Marshal(session)
	if err != nil {
		return nil, "", err
	}
	tok, _, err := s.open(ctx, PurposeRegister, userID, "", "", raw)
	if err != nil {
		return nil, "", err
	}
	opts, err := json.Marshal(creation)
	return opts, tok, err
}

// FinishKey checks the browser's attestation and stores the key under
// keyName. On a first factor it returns the person's new recovery codes.
func (s *Service) FinishKey(ctx context.Context, userID, token, keyName string, attestation []byte) (KeyView, []string, error) {
	if s.Keys == nil {
		return KeyView{}, nil, ErrKeysUnavailable
	}
	c, err := s.take(ctx, token, PurposeRegister)
	if err != nil {
		return KeyView{}, nil, err
	}
	if c.UserID != userID {
		return KeyView{}, nil, ErrChallengeExpired
	}
	var session webauthn.SessionData
	if err := json.Unmarshal(c.Session, &session); err != nil {
		return KeyView{}, nil, ErrChallengeExpired
	}
	parsed, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(attestation))
	if err != nil {
		return KeyView{}, nil, errors.New("the security key's response could not be read; try again")
	}
	u, _, err := s.keyUser(ctx, userID, "", "")
	if err != nil {
		return KeyView{}, nil, err
	}
	cred, err := s.Keys.wa.CreateCredential(u, session, parsed)
	if err != nil {
		return KeyView{}, nil, errors.New("the security key could not be verified; try again")
	}
	raw, err := json.Marshal(cred)
	if err != nil {
		return KeyView{}, nil, err
	}
	keyName = strings.TrimSpace(keyName)
	if keyName == "" {
		keyName = "Security key"
	}
	if len(keyName) > 120 {
		keyName = keyName[:120]
	}
	credID := base64.RawURLEncoding.EncodeToString(cred.ID)
	now := s.now()
	f := Factor{ID: newID(), UserID: userID, Kind: KindWebAuthn, Name: keyName, CredentialID: &credID, Credential: raw, ConfirmedAt: &now, CreatedAt: now}
	var codes []string
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.consume(ctx, c.ID); err != nil {
			return err
		}
		if err := tx.Create(&f).Error; err != nil {
			return errors.New("this security key is already registered")
		}
		codes, err = s.afterFirstFactor(ctx, tx, userID)
		return err
	})
	if err != nil {
		return KeyView{}, nil, err
	}
	return KeyView{ID: f.ID, Name: f.Name, AddedAt: now}, codes, nil
}

// KeyOptions returns the options for navigator.credentials.get for a sign-in
// or re-verification in progress (any of the person's keys will do).
func (s *Service) KeyOptions(ctx context.Context, token string) (json.RawMessage, error) {
	if s.Keys == nil {
		return nil, ErrKeysUnavailable
	}
	var c Challenge
	err := s.DB.WithContext(ctx).Where(`"token_hash" = ? AND "purpose" IN ?`, hashToken(token), []string{PurposeSignIn, PurposeStepUp}).Take(&c).Error
	if err != nil || c.ConsumedAt != nil || !s.now().Before(c.ExpiresAt) {
		return nil, ErrChallengeExpired
	}
	u, _, err := s.keyUser(ctx, c.UserID, "", "")
	if err != nil {
		return nil, err
	}
	if len(u.creds) == 0 {
		return nil, ErrNotEnrolled
	}
	assertion, session, err := s.Keys.wa.BeginLogin(u, webauthn.WithUserVerification(protocol.VerificationDiscouraged))
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}
	if err := s.DB.WithContext(ctx).Model(&Challenge{}).Where(`"id" = ?`, c.ID).Update("session", raw).Error; err != nil {
		return nil, err
	}
	return json.Marshal(assertion)
}

// useKey checks a security key's assertion against the challenge's session
// and records the key's signature counter. A counter that went backwards
// (a cloned key) is refused.
func (s *Service) useKey(ctx context.Context, c Challenge, assertion []byte) (bool, error) {
	if s.Keys == nil {
		return false, ErrKeysUnavailable
	}
	var session webauthn.SessionData
	if len(c.Session) == 0 || json.Unmarshal(c.Session, &session) != nil || session.Challenge == "" {
		return false, nil
	}
	parsed, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(assertion))
	if err != nil {
		return false, nil
	}
	u, fs, err := s.keyUser(ctx, c.UserID, "", "")
	if err != nil {
		return false, err
	}
	cred, err := s.Keys.wa.ValidateLogin(u, session, parsed)
	if err != nil || cred == nil || cred.Authenticator.CloneWarning {
		return false, nil
	}
	credID := base64.RawURLEncoding.EncodeToString(cred.ID)
	raw, err := json.Marshal(cred)
	if err != nil {
		return false, err
	}
	for _, f := range fs {
		if f.CredentialID != nil && *f.CredentialID == credID {
			err := s.DB.WithContext(ctx).Model(&Factor{}).Where(`"id" = ?`, f.ID).
				Updates(map[string]any{"credential": raw, "last_used_at": s.now()}).Error
			return err == nil, err
		}
	}
	return false, nil
}
