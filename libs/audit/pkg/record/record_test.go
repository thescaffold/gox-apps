package record

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/awesome-goose/goose/modules/sql"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/thescaffold/gox-apps/libs/audit/migrations"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("AUDIT_TEST_DB_URL")
	if dsn == "" {
		t.Skip("AUDIT_TEST_DB_URL not set; skipping integration test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	for _, m := range []sql.Migration{&migrations.CreateAuditLogs{}, &migrations.AddActorToAuditLogs{}} {
		if err := m.Run(q); err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func count(t *testing.T, db *gorm.DB, ws string) int64 {
	t.Helper()
	var n int64
	if err := db.Table("AuditLogs").Where("workspace_id = ?", ws).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func TestRecordJoinsTheCallersTransaction(t *testing.T) {
	db := testDB(t)
	ws := uuid.NewString()
	e := Entry{WorkspaceID: ws, UserID: "u1", ActorType: "user", ActorID: "u1", Service: "secrets", EntityID: "item-1", EntityName: "STRIPE_KEY", Action: "reveal", Meta: map[string]any{"scope": "project"}}

	_ = db.Transaction(func(tx *gorm.DB) error {
		if err := Record(tx, e); err != nil {
			t.Fatal(err)
		}
		return context.Canceled // roll back
	})
	if n := count(t, db, ws); n != 0 {
		t.Fatalf("rolled-back entry persisted: %d", n)
	}
	if err := db.Transaction(func(tx *gorm.DB) error { return Record(tx, e) }); err != nil {
		t.Fatal(err)
	}
	if n := count(t, db, ws); n != 1 {
		t.Fatalf("committed entries: %d, want 1", n)
	}
	if err := Record(db, Entry{WorkspaceID: ws}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("missing fields: %v", err)
	}
}

func TestAppendOnly(t *testing.T) {
	db := testDB(t)
	t.Cleanup(func() {
		db.Exec(`DROP TRIGGER IF EXISTS audit_logs_append_only ON "AuditLogs"`)
		appendOnly.Store(false)
	})
	ws := uuid.NewString()
	old := Entry{WorkspaceID: ws, Service: "identity", EntityID: "u1", Action: "role.change"}
	if err := Record(db, old); err != nil {
		t.Fatal(err)
	}
	db.Exec(`UPDATE "AuditLogs" SET created_at = ? WHERE workspace_id = ?`, time.Now().AddDate(-2, 0, 0), ws)
	if err := Record(db, Entry{WorkspaceID: ws, Service: "identity", EntityID: "u2", Action: "role.change"}); err != nil {
		t.Fatal(err)
	}

	if AppendOnly() {
		t.Fatal("append-only must be off until enabled")
	}
	for i := 0; i < 2; i++ { // idempotent
		if err := EnableAppendOnly(db); err != nil {
			t.Fatal(err)
		}
	}
	if !AppendOnly() {
		t.Fatal("append-only flag not set")
	}
	if err := db.Exec(`UPDATE "AuditLogs" SET action = 'tampered' WHERE workspace_id = ?`, ws).Error; err == nil {
		t.Fatal("an update was allowed")
	}
	if err := db.Exec(`DELETE FROM "AuditLogs" WHERE workspace_id = ?`, ws).Error; err == nil {
		t.Fatal("a delete was allowed")
	}
	// Inserts still work, and retention removes only old rows.
	if err := Record(db, Entry{WorkspaceID: ws, Service: "secrets", EntityID: "i", Action: "reveal"}); err != nil {
		t.Fatalf("insert under append-only: %v", err)
	}
	n, err := Purge(db, time.Now().AddDate(-1, 0, 0))
	if err != nil || n < 1 {
		t.Fatalf("purge: %d %v", n, err)
	}
	if c := count(t, db, ws); c != 2 {
		t.Fatalf("after purge: %d rows, want the 2 recent ones", c)
	}
}
