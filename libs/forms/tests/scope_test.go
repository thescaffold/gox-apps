package tests

import (
	"net/http"
	"os"
	"testing"

	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/forms/app"
	formsform "github.com/thescaffold/gox-apps/libs/forms/app/form"
	formsformfield "github.com/thescaffold/gox-apps/libs/forms/app/formfield"
	formsformlog "github.com/thescaffold/gox-apps/libs/forms/app/formlog"
	formsformtype "github.com/thescaffold/gox-apps/libs/forms/app/formtype"
	coreauth "github.com/thescaffold/gox-packages/libs/core/auth"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/response"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func hasAdmin(mws []types.Middleware) bool {
	for _, m := range mws {
		if _, ok := m.(*coreauth.AdminGuard); ok {
			return true
		}
	}
	return false
}

// Only GET / and the public fill-in routes (one/:id) are open; formtype and formlog writes
// need a platform admin (PLAN U26).
func TestFormsRoutesGuards(t *testing.T) {
	open := map[string]bool{"/": true, "/one/:id": true}
	adminWrites := map[types.Module]bool{formsformtype.ROUTES: true, formsformlog.ROUTES: true}
	for _, m := range []types.Module{app.ROUTES, formsform.ROUTES, formsformfield.ROUTES, formsformlog.ROUTES, formsformtype.ROUTES} {
		rs, err := m.(interface{ Routes() (types.Routes, error) }).Routes()
		if err != nil {
			t.Fatal(err)
		}
		var walk func(rs types.Routes, inherited []types.Middleware, path string)
		walk = func(rs types.Routes, inherited []types.Middleware, path string) {
			for _, r := range rs {
				chain := append(append([]types.Middleware{}, inherited...), r.Middlewares...)
				full := path + r.Path
				if r.Handler != nil {
					name := r.Handler.([]any)[1].(string)
					if !open[full] && !coreauth.RequiresSession(chain) {
						t.Errorf("%s %s is open to anonymous callers", r.Method, full)
					}
					if adminWrites[m] && (name == "Create" || name == "Update" || name == "Delete") && !hasAdmin(chain) {
						t.Errorf("%s %s should need a platform admin", r.Method, full)
					}
				}
				walk(r.Children, chain, full)
			}
		}
		walk(rs, nil, "")
	}
}

func formsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("FORMS_TEST_DB_URL")
	if dsn == "" {
		t.Skip("FORMS_TEST_DB_URL not set; skipping integration test")
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

// Workspace B can neither add a field to A's form nor see A's fields.
func TestAnotherWorkspacesFormFieldsAreUnreachable(t *testing.T) {
	db := formsTestDB(t)
	c, forms := formsformfield.NewControllerForTest((&sql.Query{}).With(&sql.Db{DB: db}))
	form := &formsform.Form{UserId: "u-a", ClientId: "c", WorkspaceId: "ws-a", TypeId: "t", Key: "scope-k", Name: "scope-form"}
	if err := forms.Insert(form); err != nil {
		t.Fatal(err)
	}
	a := ntxctx.NTXContext{UserID: "u-a", WorkspaceID: "ws-a"}
	b := ntxctx.NTXContext{UserID: "u-b", WorkspaceID: "ws-b"}
	field := formsformfield.CreateFormFieldDto{FormId: form.Id, Key: "email", Name: "Email", Type: "text"}

	if out := c.Create(&crud.CreateDto[formsformfield.CreateFormFieldDto]{Body: field, Ctx: b}); out.Code() != http.StatusNotFound {
		t.Fatalf("ws-b added a field to ws-a's form: %d", out.Code())
	}
	if out := c.Create(&crud.CreateDto[formsformfield.CreateFormFieldDto]{Body: field, Ctx: a}); out.Code() != http.StatusOK {
		t.Fatalf("ws-a could not add a field to its own form: %d", out.Code())
	}
	list := func(ctx ntxctx.NTXContext) int {
		return len(listed(t, c.List(&crud.ListDto{Queries: map[string]string{}, Ctx: ctx})))
	}
	if n := list(b); n != 0 {
		t.Fatalf("ws-b lists %d of ws-a's fields", n)
	}
	if n := list(a); n != 1 {
		t.Fatalf("ws-a lists %d fields, want 1", n)
	}
}

func listed(t *testing.T, out types.Output) []formsformfield.FormField {
	t.Helper()
	rows, ok := out.Data().(response.Envelope).Data.([]formsformfield.FormField)
	if !ok {
		t.Fatalf("list data is %T", out.Data())
	}
	return rows
}
