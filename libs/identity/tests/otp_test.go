package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/thescaffold/gox-apps/libs/identity/pkg/otp"
	"gorm.io/gorm"
)

func otpSvc(t *testing.T) (*otp.Service, *gorm.DB, *time.Time) {
	t.Helper()
	db := identityTestDB(t)
	db.Exec(`TRUNCATE "IdentityOtps"`)
	now := time.Now()
	return &otp.Service{DB: db, Now: func() time.Time { return now }, Magic: ""}, db, &now
}

func TestOtpNormalize(t *testing.T) {
	for in, want := range map[string]string{
		"  Ada@Example.COM ": "ada@example.com", "+234 801 234 5678": "+2348012345678", "08012345678": "+2348012345678",
		"2348012345678": "+2348012345678", "00234 801-234-5678": "+2348012345678", "+14155550123": "+14155550123",
	} {
		got, _, err := otp.Normalize(in)
		if err != nil || got != want {
			t.Errorf("%q: %q %v, want %q", in, got, err, want)
		}
	}
	if _, typ, _ := otp.Normalize("a@b.co"); typ != otp.Email {
		t.Error("email type")
	}
	if _, typ, _ := otp.Normalize("08012345678"); typ != otp.Phone {
		t.Error("phone type")
	}
	for _, bad := range []string{"", "ada", "ada@", "@x.com", "12345", "+0123456789", "0801234567", "+1234"} {
		if _, _, err := otp.Normalize(bad); !errors.Is(err, otp.ErrBadRef) {
			t.Errorf("%q accepted: %v", bad, err)
		}
	}
}

func TestOtpIssueCheckIsSingleUseAndStoredAsAHash(t *testing.T) {
	s, db, _ := otpSvc(t)
	code, err := s.Issue(kctx, "ada@example.com", "c1")
	if err != nil || len(code) != 6 {
		t.Fatalf("%q %v", code, err)
	}
	var plain int64
	db.Raw(`SELECT COUNT(*) FROM "IdentityOtps" WHERE code_hash = ?`, code).Scan(&plain)
	if plain != 0 {
		t.Fatal("the code is stored in the clear; only its hash may be")
	}
	if err := s.Check(kctx, "ada@example.com", "c1", code); err != nil {
		t.Fatalf("a correct code: %v", err)
	}
	if err := s.Check(kctx, "ada@example.com", "c1", code); !errors.Is(err, otp.ErrInvalid) {
		t.Fatalf("a code must work once: %v", err)
	}
	// a code is for one reference and one client
	code2, _ := s.Issue(kctx, "lin@example.com", "c1")
	if err := s.Check(kctx, "ada@example.com", "c1", code2); err == nil {
		t.Fatal("another user's code was accepted")
	}
	if err := s.Check(kctx, "lin@example.com", "c2", code2); err == nil {
		t.Fatal("a code was accepted for another client")
	}
}

func TestOtpExpiryAttemptsAndASecondCodeReplacesTheFirst(t *testing.T) {
	s, _, now := otpSvc(t)
	code, _ := s.Issue(kctx, "ada@example.com", "c1")
	wrong := "000000"
	if wrong == code {
		wrong = "111111"
	}
	for i := 0; i < 5; i++ {
		if err := s.Check(kctx, "ada@example.com", "c1", wrong); !errors.Is(err, otp.ErrInvalid) {
			t.Fatalf("guess %d: %v", i, err)
		}
	}
	if err := s.Check(kctx, "ada@example.com", "c1", code); !errors.Is(err, otp.ErrLocked) {
		t.Fatalf("after five wrong guesses even the right code is dead: %v", err)
	}

	*now = now.Add(time.Minute)
	c2, err := s.Issue(kctx, "ada@example.com", "c1")
	if err != nil {
		t.Fatal(err)
	}
	*now = now.Add(time.Minute)
	c3, _ := s.Issue(kctx, "ada@example.com", "c1")
	if c2 != c3 {
		if err := s.Check(kctx, "ada@example.com", "c1", c2); err == nil {
			t.Fatal("an older code still worked after a newer one was issued")
		}
	}
	*now = now.Add(11 * time.Minute)
	if err := s.Check(kctx, "ada@example.com", "c1", c3); !errors.Is(err, otp.ErrExpired) {
		t.Fatalf("%v", err)
	}
}

func TestOtpSendsAreThrottledAndCapped(t *testing.T) {
	s, _, now := otpSvc(t)
	if _, err := s.Issue(kctx, "ada@example.com", "c1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Issue(kctx, "ada@example.com", "c1"); !errors.Is(err, otp.ErrThrottled) {
		t.Fatalf("a second send within the gap: %v", err)
	}
	if _, err := s.Issue(kctx, "lin@example.com", "c1"); err != nil {
		t.Fatalf("throttling is per reference: %v", err)
	}
	for i := 0; i < 4; i++ {
		*now = now.Add(time.Minute)
		if _, err := s.Issue(kctx, "ada@example.com", "c1"); err != nil {
			t.Fatalf("send %d: %v", i+2, err)
		}
	}
	*now = now.Add(time.Minute)
	if _, err := s.Issue(kctx, "ada@example.com", "c1"); !errors.Is(err, otp.ErrTooMany) {
		t.Fatalf("the sixth send in an hour: %v", err)
	}
	*now = now.Add(time.Hour)
	if _, err := s.Issue(kctx, "ada@example.com", "c1"); err != nil {
		t.Fatalf("the cap resets: %v", err)
	}
}

func TestOtpMagicCodeWorksOutsideProductionOnly(t *testing.T) {
	s, _, _ := otpSvc(t)
	s.Magic = "000000"
	t.Setenv("NODE_ENV", "development")
	if err := s.Check(kctx, "ada@example.com", "c1", "000000"); err != nil {
		t.Fatalf("the sandbox magic code: %v", err)
	}
	t.Setenv("NODE_ENV", "production")
	if err := s.Check(kctx, "ada@example.com", "c1", "000000"); err == nil {
		t.Fatal("the magic code must be refused in production")
	}
}
