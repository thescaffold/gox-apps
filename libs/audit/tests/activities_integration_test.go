package tests

import (
	"os"
	"testing"

	auditlog "github.com/thescaffold/gox-apps/libs/audit/app/log"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Requires AUDIT_TEST_DB_URL to point at a Postgres database with AuditLogs
// (both migrations) already applied. Skips, not fails, when unset.
func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("AUDIT_TEST_DB_URL")
	if dsn == "" {
		t.Skip("AUDIT_TEST_DB_URL not set; skipping integration test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return db
}

// TestActivities_ReturnsAIActor is PLAN M1-04's Done criterion made literal:
// "an AI-role event appears in /activities with its actor." Inserts a real
// row with ActorType/ActorId set and no UserId (an AI-initiated event has no
// authenticated human session), then calls the actual Activities() query —
// not just constructs a struct — and checks the actor survives the round
// trip through Postgres and back.
func TestActivities_ReturnsAIActor(t *testing.T) {
	db := testDB(t)
	svc := auditlog.NewLogServiceForTest(db)
	ws := uuid.NewString()
	desc := "builder created a task"
	actorType := "ai"
	actorId := "builder"

	entry := &auditlog.Log{
		WorkspaceId: ws,
		ClientId:    "internal",
		Group:       "apps",
		Service:     "tasks",
		EntityId:    uuid.NewString(),
		EntityName:  "task",
		Action:      auditlog.LogActionTypeCreate,
		Desc:        &desc,
		ActorType:   &actorType,
		ActorId:     &actorId,
	}
	if err := svc.Create(entry); err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = auditlog.NewLogEntityForTest(db).Delete("id = ?", entry.Id)
	})

	logs, total, err := svc.Activities(1, 50)
	if err != nil {
		t.Fatalf("activities: %v", err)
	}
	if total == 0 {
		t.Fatal("expected at least one activity")
	}

	var found *auditlog.Log
	for i := range logs {
		if logs[i].Id == entry.Id {
			found = &logs[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("expected the created log to appear in Activities (desc is set, so it should pass the desc IS NOT NULL filter); got %d rows", len(logs))
	}
	if found.ActorType == nil || *found.ActorType != "ai" {
		t.Fatalf("expected ActorType %q to survive the round trip, got %v", "ai", found.ActorType)
	}
	if found.ActorId == nil || *found.ActorId != "builder" {
		t.Fatalf("expected ActorId %q to survive the round trip, got %v", "builder", found.ActorId)
	}
	if found.UserId != "" {
		t.Fatalf("expected UserId to stay empty for an AI-initiated event, got %q", found.UserId)
	}
}
