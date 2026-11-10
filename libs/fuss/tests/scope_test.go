package tests

import (
	"os"
	"testing"

	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/fuss/app"
	fusshist "github.com/thescaffold/gox-apps/libs/fuss/app/history"
	fusstok "github.com/thescaffold/gox-apps/libs/fuss/app/token"
	fusstoklog "github.com/thescaffold/gox-apps/libs/fuss/app/tokenlog"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Every fuss route but GET / needs a verified caller (PLAN U26); token-log is admin only.
func TestEveryFussRouteButHelloNeedsASession(t *testing.T) {
	var walk func(rs types.Routes, inherited []types.Middleware, path string, admin bool)
	walk = func(rs types.Routes, inherited []types.Middleware, path string, admin bool) {
		for _, r := range rs {
			chain := append(append([]types.Middleware{}, inherited...), r.Middlewares...)
			if r.Handler != nil && path+r.Path != "/" {
				if !coreauth.RequiresSession(chain) {
					t.Errorf("%s%s is open to anonymous callers", path, r.Path)
				}
				if admin && !hasAdmin(chain) {
					t.Errorf("%s%s should need a platform admin", path, r.Path)
				}
			}
			walk(r.Children, chain, path+r.Path, admin)
		}
	}
	for m, admin := range map[types.Module]bool{app.ROUTES: false, fusshist.ROUTES: false, fusstok.ROUTES: false, fusstoklog.ROUTES: true} {
		rs, err := m.(interface{ Routes() (types.Routes, error) }).Routes()
		if err != nil {
			t.Fatal(err)
		}
		walk(rs, nil, "", admin)
	}
}

func hasAdmin(mws []types.Middleware) bool {
	for _, m := range mws {
		if _, ok := m.(*coreauth.AdminGuard); ok {
			return true
		}
	}
	return false
}

func fussTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("FUSS_TEST_DB_URL")
	if dsn == "" {
		t.Skip("FUSS_TEST_DB_URL not set; skipping integration test")
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

func index(t *testing.T, svc *app.AppService, workspace, id, name string) {
	t.Helper()
	err := svc.HandleCreate(map[string]any{
		"group": "g", "service": "s", "entityName": "doc",
		"meta":    map[string]any{"id": id, "name": name},
		"context": map[string]any{"user": map[string]any{"id": "u-" + workspace}, "client": map[string]any{"id": "c"}, "workspace": map[string]any{"id": workspace}},
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Search and lookup see only the caller's workspace's index; it used to be every tenant's.
func TestSearchAndLookupStayInTheCallersWorkspace(t *testing.T) {
	svc := app.NewAppServiceForTest(fussTestDB(t))
	index(t, svc, "ws-a", "11111111-1111-1111-1111-111111111111", "zebra-scope-a")
	index(t, svc, "ws-b", "22222222-2222-2222-2222-222222222222", "zebra-scope-b")

	got, _, err := svc.Search("zebra-scope", "global", "u-ws-a", "c", "ws-a", 1, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0]["title"] != "zebra-scope-a" {
		t.Fatalf("ws-a searched %v; want only its own entity", got)
	}
	if got, _, _ := svc.Search("zebra-scope", "global", "u", "c", "", 1, 50); len(got) != 0 {
		t.Fatalf("no workspace searched %v", got)
	}
	if meta, _ := svc.Lookup("ws-a", "g", "s", "doc", "22222222-2222-2222-2222-222222222222"); meta != nil {
		t.Fatalf("ws-a looked up ws-b's entity: %v", meta)
	}
	if meta, _ := svc.Lookup("ws-b", "g", "s", "doc", "22222222-2222-2222-2222-222222222222"); meta == nil {
		t.Fatal("ws-b could not look up its own entity")
	}
}
