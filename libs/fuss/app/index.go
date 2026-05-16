package app

import (
	fusshistory "github.com/thescaffold/gox-apps/libs/fuss/app/history"
	fusstoken "github.com/thescaffold/gox-apps/libs/fuss/app/token"
	fusstokenlog "github.com/thescaffold/gox-apps/libs/fuss/app/tokenlog"
	"github.com/thescaffold/gox-packages/libs/core/events"
)

const Name = "fuss"

var AllMigrations = Migrations

var DefaultPermissions = map[string][]string{
	"guest":         {},
	"member":        {"member:create:apps:fuss:*:*:workspace", "member:update:apps:fuss:*:*:self", "member:read:apps:fuss:*:*:workspace", "member:delete:apps:fuss:*:*:self"},
	"admin":         {"admin:*:apps:fuss:*:*:workspace"},
	"global-member": {"global-member:*:apps:fuss:*:*:global"},
	"global-admin":  {"global-admin:*:apps:fuss:*:*:global"},
}

var fussAppSvc *AppService

func handleAfterInsert(_ string, payload any) {
	if fussAppSvc == nil {
		return
	}
	m, ok := payload.(map[string]any)
	if !ok {
		return
	}
	_ = fussAppSvc.HandleCreate(m)
}

func handleAfterUpdate(_ string, payload any) {
	if fussAppSvc == nil {
		return
	}
	m, ok := payload.(map[string]any)
	if !ok {
		return
	}
	_ = fussAppSvc.HandleUpdate(m)
}

func handleAfterDelete(_ string, payload any) {
	if fussAppSvc == nil {
		return
	}
	m, ok := payload.(map[string]any)
	if !ok {
		return
	}
	_ = fussAppSvc.HandleDelete(m)
}

var Subscriptions = map[string]events.EventHandler{
	"*.*.*.after-insert": handleAfterInsert,
	"*.*.*.after-update": handleAfterUpdate,
	"*.*.*.after-delete": handleAfterDelete,
}

// Top-level exports mirroring ntx-apps/libs/fuss/src/index.ts.
var (
	Entities        = []any{fusshistory.History{}, fusstoken.Token{}, fusstokenlog.TokenLog{}}
	Messages        = map[string]any{}
	UnsafeEventList = []string{}
	Paths           = []string{"translations/en/ntx/apps/fuss.yaml"}
	Jobs            = []any{}
	Crons           = []any{}
)
