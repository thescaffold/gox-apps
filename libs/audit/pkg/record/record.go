// Package record writes audit entries inside the caller's transaction and,
// when a host opts in, makes AuditLogs append-only (TRD U-S24).
//
// Record is for security-relevant events that are not plain CRUD: a secret
// revealed, a role changed, a protection bypassed. Writing it in the same
// transaction as the change means the entry exists if and only if the change
// committed.
//
// EnableAppendOnly installs a Postgres trigger that refuses UPDATE and DELETE
// on AuditLogs, and makes the log API's update and delete routes answer 403.
// Purge is the one way rows leave: the retention job, inside a transaction
// that sets audit.purge.
package record

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entry is one audit record.
type Entry struct {
	WorkspaceID string
	UserID      string // the person acting; "" for the platform
	ClientID    string
	ActorType   string // user, service, platform…
	ActorID     string
	Group       string // e.g. "apps"
	Service     string // the app, e.g. "secrets"
	EntityID    string
	EntityName  string
	Action      string // e.g. "reveal", "role.change", "delete"
	Desc        string
	Meta        any
}

// ErrInvalid: a required field is missing.
var ErrInvalid = errors.New("audit: invalid entry")

// Record inserts e into AuditLogs using tx (pass the caller's transaction).
func Record(tx *gorm.DB, e Entry) error {
	if e.WorkspaceID == "" || e.Service == "" || e.EntityID == "" || e.Action == "" {
		return fmt.Errorf("%w: workspace, service, entity id and action are required", ErrInvalid)
	}
	if e.Group == "" {
		e.Group = "apps"
	}
	if e.EntityName == "" {
		e.EntityName = e.EntityID
	}
	var meta []byte
	if e.Meta != nil {
		b, err := json.Marshal(e.Meta)
		if err != nil {
			return err
		}
		meta = b
	}
	now := time.Now().UTC()
	desc := truncate(e.Desc, 255)
	return tx.Exec(`INSERT INTO "AuditLogs"
		(id, created_at, updated_at, user_id, client_id, workspace_id, actor_type, actor_id, "group", service, entity_id, entity_name, action, "desc", meta)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.NewString(), now, now, e.UserID, e.ClientID, e.WorkspaceID, nullable(e.ActorType), nullable(e.ActorID),
		e.Group, e.Service, truncate(e.EntityID, 36), truncate(e.EntityName, 255), e.Action, nullable(desc), nullableJSON(meta)).Error
}

var appendOnly atomic.Bool

// AppendOnly reports whether EnableAppendOnly ran in this process.
func AppendOnly() bool { return appendOnly.Load() }

// EnableAppendOnly installs (idempotently) the trigger that refuses UPDATE and
// DELETE on AuditLogs, except deletes by Purge, and turns the API's update and
// delete routes off. Postgres only.
func EnableAppendOnly(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return fmt.Errorf("audit: append-only needs Postgres, not %s", db.Dialector.Name())
	}
	for _, stmt := range []string{
		`CREATE OR REPLACE FUNCTION audit_logs_append_only() RETURNS trigger AS $$
		BEGIN
			IF TG_OP = 'DELETE' AND current_setting('audit.purge', true) = 'on' THEN
				RETURN OLD;
			END IF;
			RAISE EXCEPTION 'AuditLogs is append-only' USING ERRCODE = 'insufficient_privilege';
		END $$ LANGUAGE plpgsql`,
		`DROP TRIGGER IF EXISTS audit_logs_append_only ON "AuditLogs"`,
		`CREATE TRIGGER audit_logs_append_only BEFORE UPDATE OR DELETE ON "AuditLogs"
			FOR EACH ROW EXECUTE FUNCTION audit_logs_append_only()`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}
	appendOnly.Store(true)
	return nil
}

// Purge deletes entries created before cutoff (retention, e.g. one year) and
// returns how many. It is the only delete the append-only trigger allows.
func Purge(db *gorm.DB, cutoff time.Time) (int64, error) {
	var n int64
	err := db.Transaction(func(tx *gorm.DB) error {
		if db.Dialector.Name() == "postgres" {
			if err := tx.Exec(`SET LOCAL audit.purge = 'on'`).Error; err != nil {
				return err
			}
		}
		res := tx.Exec(`DELETE FROM "AuditLogs" WHERE created_at < ?`, cutoff)
		n = res.RowsAffected
		return res.Error
	})
	return n, err
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullableJSON(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return string(b)
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n])
}
