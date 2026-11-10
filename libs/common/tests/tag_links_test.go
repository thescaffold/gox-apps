package tests

import (
	"net/http"
	"os"
	"testing"

	"github.com/awesome-goose/goose/modules/sql"
	"github.com/google/uuid"
	"github.com/thescaffold/gox-apps/libs/common/app"
	"github.com/thescaffold/gox-apps/libs/common/app/tag"
	"github.com/thescaffold/gox-apps/libs/common/app/tagtype"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// commonTestDB needs COMMON_TEST_DB_URL to point at a scratch Postgres database; it runs the
// app's real migrations. Skips, not fails, when unset.
func commonTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("COMMON_TEST_DB_URL")
	if dsn == "" {
		t.Skip("COMMON_TEST_DB_URL not set; skipping integration test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(2)
	t.Cleanup(func() { _ = sqlDB.Close() })
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	for _, m := range app.AllMigrations {
		if err := m.Run(q); err != nil {
			t.Fatalf("run migration %T: %v", m, err)
		}
	}
	return db
}

func linksOf(t *testing.T, db *gorm.DB, entityID string) []tag.Tag {
	t.Helper()
	var rows []tag.Tag
	if err := db.Where(`"entity_id" = ? AND "deleted_at" IS NULL`, entityID).Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	return rows
}

// PLAN U26: tag links are made, found and removed only inside the caller's own workspace. Before,
// the routes answered anonymous callers, and sync read every workspace's links of an entity.
func TestTagLinksStayInTheCallersWorkspace(t *testing.T) {
	db := commonTestDB(t)
	c := tagtype.NewControllerForTest(db)
	entity := uuid.NewString()
	alice := ntxctx.NTXContext{UserID: uuid.NewString(), ClientID: "c1", WorkspaceID: uuid.NewString()}
	bob := ntxctx.NTXContext{UserID: uuid.NewString(), ClientID: "c1", WorkspaceID: uuid.NewString()}
	dto := func(ctx ntxctx.NTXContext, names ...string) *tagtype.TagActionDto {
		return &tagtype.TagActionDto{GroupName: "g", ServiceName: "s", EntityName: "post", EntityId: entity, Names: names, Ctx: ctx}
	}

	if out := c.Attach(dto(alice, "news")); out.Code() != http.StatusOK {
		t.Fatalf("alice attach = %d", out.Code())
	}
	rows := linksOf(t, db, entity)
	if len(rows) != 1 || rows[0].WorkspaceId == nil || *rows[0].WorkspaceId != alice.WorkspaceID || rows[0].UserId == nil || *rows[0].UserId != alice.UserID {
		t.Fatalf("alice's link is not stamped with her session: %+v", rows)
	}

	// bob cannot remove alice's link: her tag-type is not his
	if out := c.Detach(dto(bob, "news")); out.Code() == http.StatusOK {
		t.Fatalf("bob detached alice's tag: %d", out.Code())
	}
	// bob's sync with no names sees only his own links, so it leaves alice's alone
	if out := c.Sync(dto(bob)); out.Code() != http.StatusOK {
		t.Fatalf("bob sync = %d", out.Code())
	}
	if got := linksOf(t, db, entity); len(got) != 1 {
		t.Fatalf("bob's sync touched alice's link: %d left", len(got))
	}

	// bob's own attach lands in bob's workspace, beside alice's
	if out := c.Attach(dto(bob, "news")); out.Code() != http.StatusOK {
		t.Fatalf("bob attach = %d", out.Code())
	}
	if got := linksOf(t, db, entity); len(got) != 2 {
		t.Fatalf("want one link per workspace, got %d", len(got))
	}

	// alice removes her own
	if out := c.Detach(dto(alice, "news")); out.Code() != http.StatusOK {
		t.Fatalf("alice detach = %d", out.Code())
	}
	got := linksOf(t, db, entity)
	if len(got) != 1 || *got[0].WorkspaceId != bob.WorkspaceID {
		t.Fatalf("after alice's detach want only bob's link, got %+v", got)
	}

	// no session, no links
	if out := c.Attach(dto(ntxctx.NTXContext{}, "x")); out.Code() != http.StatusForbidden {
		t.Fatalf("attach without a caller = %d, want 403", out.Code())
	}
}
