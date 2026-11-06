package migrations

import (
	"fmt"

	"github.com/awesome-goose/goose/modules/sql"

	"github.com/thescaffold/gox-apps/libs/identity/pkg/org"
)

// WorkspaceHandleRegion adds a readable, unique handle (a workspace's URL
// address) and the region its data lives in, both additive. Existing
// workspaces get a handle made from their name (made unique with -2, -3…, in
// order of creation) and the default region.
type WorkspaceHandleRegion struct{ sql.BaseMigration }

func (m *WorkspaceHandleRegion) Run(q *sql.Query) error {
	for _, s := range []string{
		`ALTER TABLE "IdentityWorkspaces" ADD COLUMN IF NOT EXISTS "handle" varchar(40)`,
		`ALTER TABLE "IdentityWorkspaces" ADD COLUMN IF NOT EXISTS "region" varchar(16)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS "uq_identity_workspaces_handle" ON "IdentityWorkspaces" (lower("handle")) WHERE "handle" IS NOT NULL`,
	} {
		if _, err := q.Exec(s); err != nil {
			return err
		}
	}
	rows, err := q.Raw(`SELECT "id", "name" FROM "IdentityWorkspaces" WHERE "handle" IS NULL ORDER BY "created_at", "id"`)
	if err != nil {
		return err
	}
	taken := map[string]bool{}
	used, err := q.Raw(`SELECT lower("handle") AS "h" FROM "IdentityWorkspaces" WHERE "handle" IS NOT NULL`)
	if err != nil {
		return err
	}
	for _, u := range used {
		taken[fmt.Sprint(u["h"])] = true
	}
	for _, r := range rows {
		base := org.HandleFor(fmt.Sprint(r["name"]))
		h := ""
		for i := 1; i < 100000; i++ {
			c := base
			if i > 1 {
				c = fmt.Sprintf("%s-%d", base, i)
			}
			if !taken[c] && org.ValidHandle(c) == nil {
				h = c
				break
			}
		}
		taken[h] = true
		if _, err := q.Exec(`UPDATE "IdentityWorkspaces" SET "handle" = ? WHERE "id" = ? AND "handle" IS NULL`, h, r["id"]); err != nil {
			return err
		}
	}
	region, _ := org.Region("")
	_, err = q.Exec(`UPDATE "IdentityWorkspaces" SET "region" = ? WHERE "region" IS NULL`, region)
	return err
}
