package tests

import (
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/awesome-goose/goose/core"
	"github.com/awesome-goose/goose/log"
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	"github.com/google/uuid"
	auditapp "github.com/thescaffold/gox-apps/libs/audit/app"
	auditlog "github.com/thescaffold/gox-apps/libs/audit/app/log"
)

type bootKernel struct{ tr types.Traverser }

func (k *bootKernel) Start(...*types.Instance) (func() error, error) {
	return func() error { return nil }, nil
}
func (k *bootKernel) Routes() []types.Route { return nil }
func (k *bootKernel) AppendRoutes(...types.Route) ([]types.Route, error) {
	return nil, nil
}
func (k *bootKernel) Container() types.Container { return k.tr.Container() }
func (k *bootKernel) Registry() types.Registry   { return k.tr.Registry() }

type auditHost struct{ cfg *sql.Config }

func (h *auditHost) Imports() []types.Module {
	return []types.Module{sql.Root(h.cfg), &auditapp.AppModule{}}
}
func (h *auditHost) Exports() []any      { return nil }
func (h *auditHost) Declarations() []any { return nil }

func sqlConfigFrom(t *testing.T, dsn string) *sql.Config {
	t.Helper()
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	port, _ := strconv.Atoi(u.Port())
	pass, _ := u.User.Password()
	return &sql.Config{
		Dialect: "postgres", Host: u.Hostname(), Port: port, User: u.User.Username(), Pass: pass,
		Name: strings.TrimPrefix(u.Path, "/"), SSLMode: "disable", Schema: "public", TimeZone: "UTC",
		Migrations: auditapp.AllMigrations,
	}
}

// PLAN U18: with audit's Subscriptions on the bus, the first after-* event
// panicked — the handlers used an AppService the container never injected
// (Declarations' throwaway instance), so its logService was nil. Boots the
// real module tree and delivers an event through the real handler.
func TestSubscriptionsUseTheInjectedService(t *testing.T) {
	dsn := os.Getenv("AUDIT_TEST_DB_URL")
	if dsn == "" {
		t.Skip("AUDIT_TEST_DB_URL not set; skipping integration test")
	}
	db := testDB(t)

	tr := core.NewTraverser()
	if err := tr.Container().Register(func() types.Log { return log.NewLog(log.AppLogChannel("")) }, "", true); err != nil {
		t.Fatal(err)
	}
	host := &auditHost{cfg: sqlConfigFrom(t, dsn)}
	if err := tr.Traverse(host); err != nil {
		t.Fatal(err)
	}
	if err := tr.OnBootHooks().ExecuteAll(func(fn func(types.Kernel) error) error { return fn(&bootKernel{tr: tr}) }); err != nil {
		t.Fatal(err)
	}
	// A second traversal (another kernel in the same process, as a host's
	// spec generator or a test does) calls Declarations again; the handlers
	// must still reach an injected service.
	if err := core.NewTraverser().Container().Register(func() types.Log { return log.NewLog(log.AppLogChannel("")) }, "", true); err != nil {
		t.Fatal(err)
	}
	_ = (&auditapp.AppModule{}).Declarations()

	ws, entity := uuid.NewString(), uuid.NewString()
	payload := map[string]any{
		"entityName": "AuditTestThings", "action": "create", "group": "g", "service": "s",
		"meta": map[string]any{"id": entity},
		"context": map[string]any{
			"user": map[string]any{"id": uuid.NewString()}, "client": map[string]any{"id": "c"},
			"workspace": map[string]any{"id": ws},
		},
	}
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("the after-insert handler panicked: %v", r)
			}
		}()
		auditapp.Subscriptions["*.*.*.after-insert"]("apps.db.audittestthings.after-insert", payload)
	}()

	var n int64
	db.Model(&auditlog.Log{}).Where("workspace_id = ? AND entity_id = ?", ws, entity).Count(&n)
	if n != 1 {
		t.Fatalf("%d audit rows for the event, want 1", n)
	}
	// Its own rows are never re-logged (no loop).
	auditapp.Subscriptions["*.*.*.after-insert"]("apps.db.auditlogs.after-insert", map[string]any{
		"entityName": "AuditLogs", "action": "create", "meta": map[string]any{"id": uuid.NewString()}, "context": payload["context"],
	})
	db.Model(&auditlog.Log{}).Where("workspace_id = ?", ws).Count(&n)
	if n != 1 {
		t.Fatalf("audit logged its own row: %d rows", n)
	}
}
