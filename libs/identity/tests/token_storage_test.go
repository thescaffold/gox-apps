package tests

import (
	"strings"
	"testing"
	"time"

	"github.com/awesome-goose/goose/modules/sql"
	"github.com/google/uuid"
	"github.com/thescaffold/gox-apps/libs/identity/app/user"
	"github.com/thescaffold/gox-apps/libs/identity/migrations"
	"github.com/thescaffold/gox-apps/libs/identity/pkg"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/tokenstore"
	"github.com/thescaffold/gox-packages/libs/core/security"
	"gorm.io/gorm"
)

// PLAN M1-03 (TRD F7 / U-S3): no plaintext refresh token and no plaintext
// provider OAuth token may sit in the database.

const storageKey = "0123456789abcdef0123456789abcdef" // 32 bytes

func countWhere(t *testing.T, db *gorm.DB, table, where string, args ...any) int64 {
	t.Helper()
	var n int64
	if err := db.Table(table).Where(where, args...).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func loginFor(t *testing.T, db *gorm.DB) (string, string, *pkg.AuthService, string) {
	t.Helper()
	auth := pkg.NewAuthServiceForTest(db)
	users := &user.UserEntity{Entity: (&sql.Entity[user.User]{}).With((&sql.Query{}).With(&sql.Db{DB: db}))}
	users.OnRegister()
	hashed, _ := security.Hash("pw-pw-pw-pw")
	ref := "tok-" + uuid.NewString()
	u := &user.User{Ref: &ref, Name: "T", Email: ref + "@example.test", Secret: &hashed}
	if err := users.Insert(u); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = users.Delete("id = ?", u.Id) })
	client := "c-" + uuid.NewString()
	access, refresh, err := auth.Login(ref, "pw-pw-pw-pw", client, uuid.NewString())
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	return access, refresh, auth, u.Id
}

func TestRefreshToken_IsStoredHashedAndStillRedeems(t *testing.T) {
	db := identityTestDB(t)
	_, refresh, auth, userID := loginFor(t, db)

	if n := countWhere(t, db, "IdentityTokens", `"token" = ?`, refresh); n != 0 {
		t.Fatal("the refresh token is in the database in plaintext")
	}
	want := tokenstore.HashRefresh(refresh)
	if n := countWhere(t, db, "IdentityTokens", `"user_id" = ? AND "type" = 'refresh' AND "token" = ?`, userID, want); n != 1 {
		t.Fatalf("expected exactly one row holding the hash, got %d", n)
	}
	if !strings.HasPrefix(want, "sha256:") || strings.Contains(want, refresh) {
		t.Fatalf("hash format: %q", want)
	}
	// It must still redeem.
	if _, err := auth.RefreshToken(refresh); err != nil {
		t.Fatalf("a valid refresh token no longer redeems: %v", err)
	}
	// Presenting the stored hash as if it were the token must not redeem.
	if _, err := auth.RefreshToken(want); err == nil {
		t.Fatal("the stored hash was accepted as a refresh token")
	}
	// A revoked token (row deleted) must not redeem.
	db.Exec(`DELETE FROM "IdentityTokens" WHERE "user_id" = ?`, userID)
	if _, err := auth.RefreshToken(refresh); err == nil {
		t.Fatal("a revoked refresh token still redeems")
	}
}

func TestMigration_HashesLegacyPlaintextRefreshTokens_Idempotently(t *testing.T) {
	db := identityTestDB(t)
	access, refresh, auth, userID := loginFor(t, db)
	_ = access

	// Put the row back the way the OLD code stored it.
	if err := db.Exec(`UPDATE "IdentityTokens" SET "token" = ? WHERE "user_id" = ? AND "type" = 'refresh'`, refresh, userID).Error; err != nil {
		t.Fatal(err)
	}
	if n := countWhere(t, db, "IdentityTokens", `"token" = ?`, refresh); n != 1 {
		t.Fatal("setup: legacy plaintext row missing")
	}

	q := (&sql.Query{}).With(&sql.Db{DB: db})
	for i := 0; i < 2; i++ { // twice: must not double-hash
		if err := (&migrations.HashRefreshTokens{}).Run(q); err != nil {
			t.Fatal(err)
		}
	}
	if n := countWhere(t, db, "IdentityTokens", `"type" = 'refresh' AND "token" NOT LIKE 'sha256:%'`); n != 0 {
		t.Fatalf("%d plaintext refresh tokens remain after the migration", n)
	}
	if n := countWhere(t, db, "IdentityTokens", `"user_id" = ? AND "token" = ?`, userID, tokenstore.HashRefresh(refresh)); n != 1 {
		t.Fatal("the migrated row must hold exactly sha256(token), not a double hash")
	}
	if _, err := auth.RefreshToken(refresh); err != nil {
		t.Fatalf("a token issued before the migration no longer redeems: %v", err)
	}
}

func TestProviderToken_SealOpenRoundTripAndTamper(t *testing.T) {
	sealed, err := tokenstore.SealProviderToken("gho_secretsecretsecret", storageKey)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(sealed, "gho_") || !strings.HasPrefix(sealed, "v1:") {
		t.Fatalf("not an encrypted envelope: %q", sealed)
	}
	if got, err := tokenstore.OpenProviderToken(sealed, storageKey); err != nil || got != "gho_secretsecretsecret" {
		t.Fatalf("round trip: %q %v", got, err)
	}
	tampered := sealed[:len(sealed)-2] + "00"
	if _, err := tokenstore.OpenProviderToken(tampered, storageKey); err == nil {
		t.Fatal("a tampered ciphertext was accepted")
	}
	if _, err := tokenstore.OpenProviderToken(sealed, "ffffffffffffffffffffffffffffffff"); err == nil {
		t.Fatal("the wrong key was accepted")
	}
	// A legacy plaintext value is returned as-is (and flagged by IsSealed).
	if tokenstore.IsSealed("gho_plain") {
		t.Fatal("plaintext reported as sealed")
	}
	if got, err := tokenstore.OpenProviderToken("gho_plain", storageKey); err != nil || got != "gho_plain" {
		t.Fatalf("legacy plaintext: %q %v", got, err)
	}
	if _, err := tokenstore.SealProviderToken("x", "short"); err == nil {
		t.Fatal("a key that is not 32 bytes was accepted")
	}
}

func TestMigration_EncryptsLegacyProviderTokens_AndFailsLoudlyWithoutAKey(t *testing.T) {
	db := identityTestDB(t)
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	id := uuid.NewString()
	if err := db.Exec(`INSERT INTO "IdentityProviders" (id, created_at, updated_at, user_id, type, reference, token)
		VALUES (?, now(), now(), 'u', 'github', ?, 'gho_legacy_plaintext')`, id, id).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM "IdentityProviders" WHERE id = ?`, id) })

	// No key configured but plaintext exists: refuse to "succeed" silently.
	if err := (&migrations.EncryptProviderTokens{Key: ""}).Run(q); err == nil {
		t.Fatal("the migration reported success while plaintext tokens remain and no key is configured")
	}

	m := &migrations.EncryptProviderTokens{Key: storageKey}
	for i := 0; i < 2; i++ { // idempotent
		if err := m.Run(q); err != nil {
			t.Fatal(err)
		}
	}
	var stored string
	db.Raw(`SELECT token FROM "IdentityProviders" WHERE id = ?`, id).Scan(&stored)
	if strings.Contains(stored, "gho_legacy") || !strings.HasPrefix(stored, "v1:") {
		t.Fatalf("stored token is still readable: %q", stored)
	}
	if got, err := tokenstore.OpenProviderToken(stored, storageKey); err != nil || got != "gho_legacy_plaintext" {
		t.Fatalf("after migration: %q %v", got, err)
	}
}

func TestMigration_ScrubsAccessTokensFromProviderLogs(t *testing.T) {
	db := identityTestDB(t)
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	id := uuid.NewString()
	if err := db.Exec(`INSERT INTO "IdentityProviderLogs" (id, created_at, updated_at, provider_id, event, request, response)
		VALUES (?, now(), now(), 'p', 'profile', '{"accessToken":"gho_in_a_log","keep":"me"}', '{"email":"a@b.c"}')`, id).Error; err != nil {
		t.Skipf("provider log schema differs: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM "IdentityProviderLogs" WHERE id = ?`, id) })
	if err := (&migrations.ScrubProviderLogTokens{}).Run(q); err != nil {
		t.Fatal(err)
	}
	var req string
	db.Raw(`SELECT request::text FROM "IdentityProviderLogs" WHERE id = ?`, id).Scan(&req)
	if strings.Contains(req, "gho_in_a_log") || !strings.Contains(req, "keep") {
		t.Fatalf("provider log request after scrub: %s", req)
	}
	_ = time.Second
}

// The OTP / OAuth sign-in path issues tokens through a different method
// (IssueTokensByHandle) than password Login; it must store the hash too.
func TestIssueTokensByHandle_StoresRefreshHashed(t *testing.T) {
	db := identityTestDB(t)
	_, _, auth, userID := loginFor(t, db)
	users := &user.UserEntity{Entity: (&sql.Entity[user.User]{}).With((&sql.Query{}).With(&sql.Db{DB: db}))}
	users.OnRegister()
	u, err := users.First("id = ?", userID)
	if err != nil {
		t.Fatal(err)
	}
	_, refresh, err := auth.IssueTokensByHandle(*u.Ref, "c-"+uuid.NewString(), uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	if n := countWhere(t, db, "IdentityTokens", `"token" = ?`, refresh); n != 0 {
		t.Fatal("IssueTokensByHandle stored the refresh token in plaintext")
	}
	if _, err := auth.RefreshToken(refresh); err != nil {
		t.Fatalf("it must still redeem: %v", err)
	}
}
