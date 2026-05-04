package app

import "github.com/thescaffold/gox-packages-core/events"

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
