package migrations

import "github.com/awesome-goose/goose/modules/sql"

// AlignPolylogEventEntity is the Phase 4 schema-alignment migration for
// polylog. It adds `entity` on PolylogEvents to mirror TS
// ntx-apps/libs/polylog/src/api/event/entities/event.entity.ts line 29
// (`@Column({name: 'entity'})` for the entityName field). gox originally
// stored this as `entity_name`. Both columns coexist after this migration —
// new writes populate `entity`; the existing `entity_name` index stays in
// place. Idempotent ALTER TABLE ADD COLUMN IF NOT EXISTS.
type AlignPolylogEventEntity struct{ sql.BaseMigration }

func (m *AlignPolylogEventEntity) Run(q *sql.Query) error {
	for _, stmt := range []string{
		`ALTER TABLE "PolylogEvents" ADD COLUMN IF NOT EXISTS "entity" varchar(255) DEFAULT NULL`,
		`CREATE INDEX IF NOT EXISTS "idx_polylog_events_entity" ON "PolylogEvents" ("entity")`,
	} {
		if _, err := q.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
