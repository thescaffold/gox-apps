package tests

import (
	"os"
	"testing"

	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/blobs/app"
	"github.com/thescaffold/gox-apps/libs/blobs/app/file"
	"github.com/thescaffold/gox-apps/libs/blobs/app/page"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Every blobs route but GET / needs a verified caller, and the resources are scoped (PLAN U26).
func TestEveryBlobsRouteButHelloNeedsASession(t *testing.T) {
	var all types.Routes
	for _, m := range []types.Module{app.ROUTES, file.ROUTES, page.ROUTES} {
		rs, err := m.(interface{ Routes() (types.Routes, error) }).Routes()
		if err != nil {
			t.Fatal(err)
		}
		all = append(all, rs...)
	}
	var walk func(rs types.Routes, inherited []types.Middleware, path string)
	walk = func(rs types.Routes, inherited []types.Middleware, path string) {
		for _, r := range rs {
			chain := append(append([]types.Middleware{}, inherited...), r.Middlewares...)
			if r.Handler != nil && path+r.Path != "/" && !coreauth.RequiresSession(chain) {
				t.Errorf("%s%s is open to anonymous callers", path, r.Path)
			}
			walk(r.Children, chain, path+r.Path)
		}
	}
	walk(all, nil, "")
}

func blobsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("BLOBS_TEST_DB_URL")
	if dsn == "" {
		t.Skip("BLOBS_TEST_DB_URL not set; skipping integration test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		if s, err := db.DB(); err == nil {
			_ = s.Close()
		}
	})
	q := (&sql.Query{}).With(&sql.Db{DB: db})
	for _, m := range app.AllMigrations {
		if err := m.Run(q); err != nil {
			t.Fatalf("run migration %T: %v", m, err)
		}
	}
	return db
}

// One workspace can neither write pages into, nor download, another workspace's file.
func TestAnotherWorkspacesFileIsUnreachable(t *testing.T) {
	svc := app.NewAppServiceForTest(blobsTestDB(t))
	owner, err := svc.Upload(app.InitInput{UserId: "u-a", ClientId: "c", WorkspaceId: "ws-a", Type: "doc", Name: "scope-test.txt", Size: 5, Mime: "text/plain"}, []string{"aGVsbG8="})
	if err != nil {
		t.Fatal(err)
	}
	id := owner.File.File.Id

	if data, _, err := svc.Download("ws-a", id); err != nil || string(data) != "hello" {
		t.Fatalf("owner download = %q, %v", data, err)
	}
	if _, _, err := svc.Download("ws-b", id); err != app.ErrFileNotFound {
		t.Fatalf("another workspace downloaded the file: %v", err)
	}
	if _, _, err := svc.Download("ws-b", "scope-test.txt"); err != app.ErrFileNotFound {
		t.Fatalf("another workspace downloaded the file by name: %v", err)
	}
	if _, _, err := svc.Download("", id); err != app.ErrFileNotFound {
		t.Fatalf("no workspace downloaded the file: %v", err)
	}
	if _, err := svc.Batch("ws-b", []app.BatchInput{{FileId: id, Index: 7, Raw: []byte("x")}}); err != app.ErrFileNotFound {
		t.Fatalf("another workspace wrote a page: %v", err)
	}
	if _, _, err := svc.Download("ws-a", "00000000-0000-0000-0000-000000000000"); err != app.ErrFileNotFound {
		t.Fatalf("a missing file must be not-found, got %v", err)
	}
}
